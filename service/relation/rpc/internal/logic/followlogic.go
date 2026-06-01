package logic

import (
	"context"
	"fmt"
	"strconv"
	"time"

	errutil "SChill/common/error"
	"SChill/common/mq"
	commonredis "SChill/common/redis"
	"SChill/service/relation/rpc/internal/model"
	"SChill/service/relation/rpc/internal/svc"
	"SChill/service/relation/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type FollowLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFollowLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FollowLogic {
	return &FollowLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *FollowLogic) Follow(in *pb.FollowReq) (*pb.FollowResp, error) {
	if in.UserId == in.TargetUserId {
		return nil, errutil.RpcBusinessError(errutil.ErrCannotFollowSelf)
	}

	if err := ensureRelationUsersExist(l.ctx, l.svcCtx.UserRpc, in.UserId, in.TargetUserId); err != nil {
		logx.Errorf("validate relation users failed: userId=%d targetUserId=%d err=%v", in.UserId, in.TargetUserId, err)
		return nil, err
	}

	var existing model.Following
	err := l.svcCtx.DB.WithContext(l.ctx).
		Where("user_id = ? AND follow_id = ?", in.UserId, in.TargetUserId).
		First(&existing).Error
	if err == nil {
		return nil, errutil.RpcBusinessError(errutil.ErrAlreadyFollowed)
	}
	if err != gorm.ErrRecordNotFound {
		logx.Errorf("query existing follow failed: %v", err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	var isMutual bool
	err = l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		var reverseFollowing model.Following
		err = tx.WithContext(l.ctx).
			Where("user_id = ? AND follow_id = ?", in.TargetUserId, in.UserId).
			First(&reverseFollowing).Error
		isMutual = err == nil

		following := model.Following{
			UserID:   in.UserId,
			FollowID: in.TargetUserId,
			IsMutual: isMutual,
		}
		if err := tx.WithContext(l.ctx).Create(&following).Error; err != nil {
			return err
		}

		if isMutual {
			if err := tx.WithContext(l.ctx).
				Model(&reverseFollowing).
				Update("is_mutual", true).Error; err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		logx.Errorf("create follow relation failed: %v", err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	now := time.Now().Unix()
	shouldSend := false
	cacheResult, err := l.svcCtx.Redis.EvalCtx(
		l.ctx,
		followRelationLua,
		[]string{
			relationFollowingKey(in.UserId),
			relationFollowersKey(in.TargetUserId),
			relationFollowingEmptyKey(in.UserId),
			relationFollowersEmptyKey(in.TargetUserId),
		},
		strconv.FormatUint(in.TargetUserId, 10),
		strconv.FormatUint(in.UserId, 10),
		now,
		commonredis.RelationCacheExpire,
	)
	if err != nil {
		logx.Errorf("update follow cache failed: userId=%d targetUserId=%d err=%v", in.UserId, in.TargetUserId, err)
	} else if cacheChanged, ok := cacheResult.(int64); ok {
		shouldSend = cacheChanged == 1
	}

	if shouldSend {
		followMsg := mq.UserFollowedMessage{
			FollowerID:  in.UserId,
			FollowingID: in.TargetUserId,
		}
		if err := l.svcCtx.KafkaProducer.SendEvent(
			l.svcCtx.Config.KqPusherConf.TopicFollowed,
			fmt.Sprintf("%d:%d", in.UserId, in.TargetUserId),
			"user.followed",
			"relation-rpc",
			"relation",
			fmt.Sprintf("%d:%d", in.UserId, in.TargetUserId),
			followMsg,
		); err != nil {
			logx.Errorf("send follow kafka message failed: %v", err)
		}
	}

	return &pb.FollowResp{
		Success:  true,
		IsMutual: isMutual,
	}, nil
}
