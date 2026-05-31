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
	redisMiss := false

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
			redisMiss = true
			for _, postID := range chunk {
				collectionMap[postID] = false
			}
			continue
		}

		for relationKey, isCollected := range statuses {
			collectionMap[keyToPostID[relationKey]] = isCollected
		}
	}

	// If any Redis chunk failed, fall back to DB for ALL post IDs
	if redisMiss {
		dbMap, dbErr := l.svcCtx.PostCollectionDAO.BatchCheckStatus(l.ctx, in.UserId, postIDs)
		if dbErr != nil {
			logx.Errorf("DB batch collection status fallback failed: userId=%d err=%v", in.UserId, dbErr)
			return &pb.BatchCheckPostCollectionResp{
				CollectionStatus: collectionMap,
			}, nil
		}
		return &pb.BatchCheckPostCollectionResp{
			CollectionStatus: dbMap,
		}, nil
	}

	return &pb.BatchCheckPostCollectionResp{
		CollectionStatus: collectionMap,
	}, nil
}
