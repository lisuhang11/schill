package logic

import (
	"context"

	errutil "SChill/common/error"
	userpb "SChill/service/user/rpc/pb"
	"SChill/service/user/rpc/usercenter"
)

type relationUserProfile struct {
	UserID   uint64
	Username string
	Avatar   string
}

func batchLoadRelationUsers(ctx context.Context, userRpc usercenter.UserCenter, userIDs []uint64) (map[uint64]relationUserProfile, error) {
	resp, err := userRpc.BatchGetUserBasicInfo(ctx, &userpb.BatchGetUserBasicInfoReq{
		UserIds: uniqueUint64s(userIDs),
	})
	if err != nil {
		return nil, err
	}

	result := make(map[uint64]relationUserProfile, len(resp.GetUsers()))
	for _, user := range resp.GetUsers() {
		if user == nil || user.GetId() == 0 {
			continue
		}
		username := user.GetNickname()
		if username == "" {
			username = user.GetUsername()
		}
		result[user.GetId()] = relationUserProfile{
			UserID:   user.GetId(),
			Username: username,
			Avatar:   user.GetAvatar(),
		}
	}

	return result, nil
}

func ensureRelationUsersExist(ctx context.Context, userRpc usercenter.UserCenter, userIDs ...uint64) error {
	resp, err := userRpc.BatchGetUserBasicInfo(ctx, &userpb.BatchGetUserBasicInfoReq{
		UserIds: uniqueUint64s(userIDs),
	})
	if err != nil {
		return err
	}
	if len(resp.GetUsers()) != len(uniqueUint64s(userIDs)) {
		return errutil.RpcBusinessError(errutil.ErrUserNotExist)
	}
	return nil
}

func uniqueUint64s(values []uint64) []uint64 {
	seen := make(map[uint64]struct{}, len(values))
	result := make([]uint64, 0, len(values))
	for _, value := range values {
		if value == 0 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
