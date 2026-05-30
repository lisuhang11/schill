package logic

import (
	"context"
	"fmt"
	"strconv"

	errutil "SChill/common/error"
	"SChill/common/redis"
	"SChill/service/interaction/rpc/internal/svc"
	"SChill/service/interaction/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type BatchCheckPostCollectionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBatchCheckPostCollectionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchCheckPostCollectionLogic {
	return &BatchCheckPostCollectionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *BatchCheckPostCollectionLogic) BatchCheckPostCollection(in *pb.BatchCheckPostCollectionReq) (*pb.BatchCheckPostCollectionResp, error) {
	postIDs := uniquePostIDs(in.PostIds)
	if len(postIDs) > maxBatchStatusPostIDs {
		return nil, errutil.RpcBusinessError(errutil.ErrInvalidParams)
	}
	if len(postIDs) == 0 || in.UserId == 0 {
		return &pb.BatchCheckPostCollectionResp{
			CollectionStatus: make(map[uint64]bool),
		}, nil
	}

	userIDStr := strconv.FormatUint(in.UserId, 10)
	collectionMap := make(map[uint64]bool, len(postIDs))

	for start := 0; start < len(postIDs); start += batchStatusChunkSize {
		end := start + batchStatusChunkSize
		if end > len(postIDs) {
			end = len(postIDs)
		}

		chunk := postIDs[start:end]
		checks := make(map[string]string, len(chunk))
		keyToPostID := make(map[string]uint64, len(chunk))
		for _, postID := range chunk {
			relationKey := fmt.Sprintf("%s%d", redis.PostCollectionRelationKey, postID)
			checks[relationKey] = userIDStr
			keyToPostID[relationKey] = postID
		}

		statuses, err := l.svcCtx.RedisClient.PipelineSIsMember(l.ctx, checks)
		if err != nil {
			logx.Errorf("Redis pipeline collection status query failed: chunk=%v err=%v", chunk, err)
			for _, postID := range chunk {
				collectionMap[postID] = false
			}
			continue
		}

		for relationKey, isCollected := range statuses {
			collectionMap[keyToPostID[relationKey]] = isCollected
		}
	}

	return &pb.BatchCheckPostCollectionResp{
		CollectionStatus: collectionMap,
	}, nil
}
