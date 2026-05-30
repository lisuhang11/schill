package logic

import (
	"context"
	"fmt"
	"strconv"
	"time"

	errutil "SChill/common/error"
	"SChill/common/mq"
	"SChill/common/redis"
	"SChill/service/interaction/rpc/internal/svc"
	"SChill/service/interaction/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type UncollectPostLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUncollectPostLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UncollectPostLogic {
	return &UncollectPostLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UncollectPostLogic) UncollectPost(in *pb.UncollectPostReq) (*pb.UncollectPostResp, error) {
	if in.UserId == 0 || in.PostId == 0 {
		return nil, errutil.RpcBusinessError(errutil.ErrInvalidParams)
	}

	relationKey := fmt.Sprintf("%s%d", redis.PostCollectionRelationKey, in.PostId)
	countKey := fmt.Sprintf("%s%d", redis.PostCollectionCountKey, in.PostId)
	timestamp := time.Now().Unix()

	result, err := l.svcCtx.RedisClient.Eval(l.ctx, l.svcCtx.LuaScripts.PostUncollect,
		[]string{relationKey, countKey},
		strconv.FormatUint(in.UserId, 10),
	)
	if err != nil {
		logx.Errorf("Redis Lua script failed: %v", err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	resultArray, ok := result.([]interface{})
	if !ok || len(resultArray) != 2 {
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	status, _ := resultArray[0].(int64)
	if status == 0 {
		go func() {
			msg := mq.PostUncollectMessage{
				PostID:    in.PostId,
				UserID:    in.UserId,
				Timestamp: timestamp,
			}
			if err := l.svcCtx.KafkaProducer.SendEvent(
				l.svcCtx.Config.KqProducerConf.TopicPostUncollect,
				strconv.FormatUint(in.PostId, 10),
				"interaction.post.uncollected",
				"interaction-rpc",
				"post",
				strconv.FormatUint(in.PostId, 10),
				msg,
			); err != nil {
				logx.Errorf("send post uncollect event failed: %v", err)
			}
		}()
	}

	return &pb.UncollectPostResp{
		Success:     true,
		IsCollected: false,
	}, nil
}
