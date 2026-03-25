package logic

import (
	errutil "SChill/common/error"
	"SChill/service/user/api/internal/svc"
	"SChill/service/user/api/internal/types"
	"SChill/service/user/rpc/pb"
	"context"
	"encoding/json"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateUserProfileInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateUserProfileInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserProfileInfoLogic {
	return &UpdateUserProfileInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateUserProfileInfoLogic) UpdateUserProfileInfo(req *types.UpdateUserProfileInfoReq) (resp *types.UpdateUserProfileInfoResp, err error) {
	logx.Infof("收到更新用户资料请求: birthday=[%s], gender=[%d], website=[%s]",
		req.UserProfile.Birthday, req.UserProfile.Gender, req.UserProfile.Website)

	logx.Infof("Context keys: %v", l.ctx)

	userId, err := l.getUserIdFromContext()
	logx.Infof("从Context获取到的userId: [%d]", userId)

	if err != nil {
		return &types.UpdateUserProfileInfoResp{
			Code: errutil.ErrUnauthorized,
			Msg:  errutil.GetCodeMessage(errutil.ErrUnauthorized),
		}, nil
	}

	if err := req.Validate(); err != nil {
		logx.Errorf("更新用户资料参数校验失败: %v", err)
		return &types.UpdateUserProfileInfoResp{
			Code: errutil.ErrInvalidParams,
			Msg:  errutil.GetCodeMessage(errutil.ErrInvalidParams),
		}, nil
	}

	rpcReq := &pb.UpdateUserProfileInfoReq{
		UserProfile: &pb.UserProfile{
			UserId:    userId,
			Gender:    int64(req.UserProfile.Gender),
			Birthday:  req.UserProfile.Birthday,
			Signature: req.UserProfile.Signature,
			Location:  req.UserProfile.Location,
			Website:   req.UserProfile.Website,
			Company:   req.UserProfile.Company,
			JobTitle:  req.UserProfile.JobTitle,
			Education: req.UserProfile.Education,
		},
	}

	logx.Infof("准备调用RPC，设置的userId=[%d], gender=[%d]",
		rpcReq.UserProfile.UserId, rpcReq.UserProfile.Gender)

	rpcResp, err := l.svcCtx.UserRpc.UpdateUserProfileInfo(l.ctx, rpcReq)
	if err != nil {
		logx.Errorf("调用 RPC UpdateUserProfileInfo 失败: %v", err)
		code, msg := errutil.ParseRpcError(err)
		return &types.UpdateUserProfileInfoResp{
			Code: code,
			Msg:  msg,
		}, nil
	}

	apiProfile := types.UserProfileInfo{
		UserId:    rpcResp.UserProfile.UserId,
		Gender:    int(rpcResp.UserProfile.Gender),
		Birthday:  rpcResp.UserProfile.Birthday,
		Signature: rpcResp.UserProfile.Signature,
		Location:  rpcResp.UserProfile.Location,
		Website:   rpcResp.UserProfile.Website,
		Company:   rpcResp.UserProfile.Company,
		JobTitle:  rpcResp.UserProfile.JobTitle,
		Education: rpcResp.UserProfile.Education,
	}

	return &types.UpdateUserProfileInfoResp{
		Code:        errutil.Success,
		Msg:         errutil.GetCodeMessage(errutil.Success),
		UserProfile: apiProfile,
	}, nil
}

func (l *UpdateUserProfileInfoLogic) getUserIdFromContext() (uint64, error) {
	logx.Infof("尝试从Context获取userId...")

	userIdVal := l.ctx.Value("userId")
	logx.Infof("userId值: %v, 类型: %T", userIdVal, userIdVal)

	switch v := userIdVal.(type) {
	case uint64:
		logx.Infof("从userId key获取到(uint64): %d", v)
		return v, nil
	case int64:
		logx.Infof("从userId key获取到(int64): %d", v)
		return uint64(v), nil
	case int:
		logx.Infof("从userId key获取到(int): %d", v)
		return uint64(v), nil
	case float64:
		logx.Infof("从userId key获取到(float64): %f", v)
		return uint64(v), nil
	case json.Number:
		logx.Infof("从userId key获取到(json.Number): %s", v)
		if userId, err := v.Int64(); err == nil {
			return uint64(userId), nil
		}
		if userId, err := v.Float64(); err == nil {
			return uint64(userId), nil
		}
	}

	logx.Errorf("无法从Context获取userId")
	return 0, nil
}
