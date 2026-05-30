package logic

import (
	"context"

	errutil "SChill/common/error"
	"SChill/service/user/rpc/internal/model"
	"SChill/service/user/rpc/internal/svc"
	"SChill/service/user/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

const maxBatchUserBasicInfoSize = 200

type BatchGetUserBasicInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBatchGetUserBasicInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchGetUserBasicInfoLogic {
	return &BatchGetUserBasicInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *BatchGetUserBasicInfoLogic) BatchGetUserBasicInfo(in *pb.BatchGetUserBasicInfoReq) (*pb.BatchGetUserBasicInfoResp, error) {
	userIDs := uniqueUint64s(in.GetUserIds())
	if len(userIDs) == 0 {
		return &pb.BatchGetUserBasicInfoResp{Users: []*pb.UserBasicInfo{}}, nil
	}
	if len(userIDs) > maxBatchUserBasicInfoSize {
		return nil, errutil.RpcBusinessError(errutil.ErrInvalidParams)
	}

	userMap := make(map[uint64]*pb.UserBasicInfo, len(userIDs))
	var missedIDs []uint64

	for _, userID := range userIDs {
		cacheKey := buildUserBasicInfoCacheKey(userID)
		var cached pb.UserBasicInfo
		err := l.svcCtx.Cache.GetCtx(l.ctx, cacheKey, &cached)
		if err == nil {
			if cached.Id > 0 {
				userMap[userID] = &cached
			}
			continue
		}

		if l.svcCtx.Cache.IsNotFound(err) {
			missedIDs = append(missedIDs, userID)
			continue
		}

		logx.Errorf("get cached user basic info failed: userId=%d err=%v", userID, err)
		missedIDs = append(missedIDs, userID)
	}

	if len(missedIDs) > 0 {
		var users []model.User
		if err := l.svcCtx.DB.WithContext(l.ctx).
			Where("id IN ? AND deleted_at IS NULL", missedIDs).
			Find(&users).Error; err != nil {
			logx.Errorf("batch get users failed: userIds=%v err=%v", missedIDs, err)
			return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
		}

		foundUserMap := make(map[uint64]*model.User, len(users))
		for i := range users {
			foundUserMap[users[i].ID] = &users[i]
		}

		for _, userID := range missedIDs {
			cacheKey := buildUserBasicInfoCacheKey(userID)
			
			if user, ok := foundUserMap[userID]; ok {
				pbUser := &pb.UserBasicInfo{
					Id:       user.ID,
					Username: user.Username,
					Nickname: user.Username,
					Avatar:   user.Avatar,
					Status:   int32(user.Status),
				}
				userMap[userID] = pbUser
				
				if err := l.svcCtx.Cache.SetCtx(l.ctx, cacheKey, pbUser); err != nil {
					logx.Errorf("cache user basic info failed: userId=%d err=%v", userID, err)
				}
			} else {
				emptyUser := &pb.UserBasicInfo{}
				if err := l.svcCtx.Cache.SetCtx(l.ctx, cacheKey, emptyUser); err != nil {
					logx.Errorf("cache empty user basic info failed: userId=%d err=%v", userID, err)
				}
			}
		}
	}

	resp := &pb.BatchGetUserBasicInfoResp{
		Users: make([]*pb.UserBasicInfo, 0, len(userIDs)),
	}
	for _, userID := range userIDs {
		if item, ok := userMap[userID]; ok {
			resp.Users = append(resp.Users, item)
		}
	}

	return resp, nil
}
