package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	errutil "SChill/common/error"
	"SChill/common/mq"
	commonredis "SChill/common/redis"
	"SChill/service/relation/rpc/internal/model"
	"SChill/service/relation/rpc/internal/svc"
	"SChill/service/relation/rpc/pb"

	"github.com/IBM/sarama"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type UnfollowLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUnfollowLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnfollowLogic {
	return &UnfollowLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UnfollowLogic) Unfollow(in *pb.UnfollowReq) (*pb.UnfollowResp, error) {
	if err := ensureRelationUsersExist(l.ctx, l.svcCtx.UserRpc, in.UserId, in.TargetUserId); err != nil {
		logx.Errorf("validate relation users failed before unfollow: userId=%d targetUserId=%d err=%v", in.UserId, in.TargetUserId, err)
		return nil, err
	}

	var following model.Following
	err := l.svcCtx.DB.WithContext(l.ctx).
		Where("user_id = ? AND follow_id = ?", in.UserId, in.TargetUserId).
		First(&following).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errutil.RpcBusinessError(errutil.ErrNotFollowing)
		}
		logx.Errorf("query follow relation failed: %v", err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	wasMutual := following.IsMutual
	err = l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.WithContext(l.ctx).Delete(&following).Error; err != nil {
			return err
		}
		if wasMutual {
			if err := tx.WithContext(l.ctx).
				Model(&model.Following{}).
				Where("user_id = ? AND follow_id = ?", in.TargetUserId, in.UserId).
				Update("is_mutual", false).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		logx.Errorf("delete follow relation failed: %v", err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	shouldSend := false
	cacheResult, err := l.svcCtx.Redis.EvalCtx(
		l.ctx,
		unfollowRelationLua,
		[]string{
			relationFollowingKey(in.UserId),
			relationFollowersKey(in.TargetUserId),
			relationFollowingEmptyKey(in.UserId),
			relationFollowersEmptyKey(in.TargetUserId),
		},
		strconv.FormatUint(in.TargetUserId, 10),
		strconv.FormatUint(in.UserId, 10),
		commonredis.RelationCacheExpire,
	)
	if err != nil {
		logx.Errorf("update unfollow cache failed: userId=%d targetUserId=%d err=%v", in.UserId, in.TargetUserId, err)
	} else if cacheChanged, ok := cacheResult.(int64); ok {
		shouldSend = cacheChanged == 1
	}

	if shouldSend {
		unfollowMsg := mq.UserUnfollowedMessage{
			FollowerID:  in.UserId,
			FollowingID: in.TargetUserId,
		}
		msgBytes, err := json.Marshal(unfollowMsg)
		if err != nil {
			logx.Errorf("marshal unfollow message failed: %v", err)
		} else {
			_, _, err := l.svcCtx.KafkaProducer.SendMessage(&sarama.ProducerMessage{
				Topic: l.svcCtx.Config.KqPusherConf.TopicUnfollowed,
				Key:   sarama.StringEncoder(fmt.Sprintf("%d:%d", in.UserId, in.TargetUserId)),
				Value: sarama.StringEncoder(string(msgBytes)),
			})
			if err != nil {
				logx.Errorf("send unfollow kafka message failed: %v", err)
			}
		}
	}

	return &pb.UnfollowResp{
		Success: true,
	}, nil
}
