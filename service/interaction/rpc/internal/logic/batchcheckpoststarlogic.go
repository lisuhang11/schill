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

type BatchCheckPostStarLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBatchCheckPostStarLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchCheckPostStarLogic {
	return &BatchCheckPostStarLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *BatchCheckPostStarLogic) BatchCheckPostStar(in *pb.BatchCheckPostStarReq) (*pb.BatchCheckPostStarResp, error) {
	postIDs := uniquePostIDs(in.PostIds)
	if len(postIDs) > maxBatchStatusPostIDs {
		return nil, errutil.RpcBusinessError(errutil.ErrInvalidParams)
	}
	if len(postIDs) == 0 || in.UserId == 0 {
		return &pb.BatchCheckPostStarResp{
			StarStatus: make(map[uint64]bool),
		}, nil
	}

	userIDStr := strconv.FormatUint(in.UserId, 10)
	starMap := make(map[uint64]bool, len(postIDs))
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
			relationKey := fmt.Sprintf("%s%d", redis.PostLikeRelationKey, postID)
			checks[relationKey] = userIDStr
			keyToPostID[relationKey] = postID
		}

		statuses, err := l.svcCtx.RedisClient.PipelineSIsMember(l.ctx, checks)
		if err != nil {
			logx.Errorf("Redis pipeline star status query failed: chunk=%v err=%v", chunk, err)
			redisMiss = true
			for _, postID := range chunk {
				starMap[postID] = false
			}
			continue
		}

		for relationKey, isStarred := range statuses {
			starMap[keyToPostID[relationKey]] = isStarred
		}
	}

	// If any Redis chunk failed, fall back to DB for ALL post IDs
	// (partial fallback is unsafe — we need consistent results)
	if redisMiss {
		dbMap, dbErr := l.svcCtx.PostStarDAO.BatchCheckStatus(l.ctx, in.UserId, postIDs)
		if dbErr != nil {
			logx.Errorf("DB batch star status fallback failed: userId=%d err=%v", in.UserId, dbErr)
			return &pb.BatchCheckPostStarResp{
				StarStatus: starMap,
			}, nil
		}
		return &pb.BatchCheckPostStarResp{
			StarStatus: dbMap,
		}, nil
	}

	return &pb.BatchCheckPostStarResp{
		StarStatus: starMap,
	}, nil
}
