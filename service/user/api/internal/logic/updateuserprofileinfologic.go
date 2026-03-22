package logic

import (
	"SChill/common/errorcode"
	"SChill/service/user/api/internal/svc"
	"SChill/service/user/api/internal/types"
	"SChill/service/user/rpc/pb"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	// 1. 从 JWT 中获取当前用户 ID（假设中间件已将 userId 存入 ctx）
	userId, ok := l.ctx.Value("userId").(uint64)
	if !ok || userId == 0 {
		return &types.UpdateUserProfileInfoResp{
			Code: errorcode.ErrUnauthorized,
			Msg:  errorcode.GetCodeMessage(errorcode.ErrUnauthorized),
		}, nil
	}

	// 2. 参数校验（使用 validator）
	if err := req.Validate(); err != nil {
		logx.Infof("更新用户资料参数校验失败: %v", err)
		return &types.UpdateUserProfileInfoResp{
			Code: errorcode.ErrInvalidParams,
			Msg:  errorcode.GetCodeMessage(errorcode.ErrInvalidParams),
		}, nil
	}

	// 3. 构造 RPC 请求（填充 userId，客户端无需传入）
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

	// 4. 调用 RPC 更新
	rpcResp, err := l.svcCtx.UserRpc.UpdateUserProfileInfo(l.ctx, rpcReq)
	if err != nil {
		logx.Errorf("调用 RPC UpdateUserProfileInfo 失败: %v", err)

		// 解析 gRPC 错误状态，映射为业务错误码
		st, _ := status.FromError(err)
		switch st.Code() {
		case codes.InvalidArgument:
			return &types.UpdateUserProfileInfoResp{
				Code: errorcode.ErrInvalidParams,
				Msg:  errorcode.GetCodeMessage(errorcode.ErrInvalidParams),
			}, nil
		case codes.NotFound:
			// 用户不存在（理论上 userId 来自 token，不应该发生）
			return &types.UpdateUserProfileInfoResp{
				Code: errorcode.ErrUserNotExist,
				Msg:  errorcode.GetCodeMessage(errorcode.ErrUserNotExist),
			}, nil
		case codes.Internal:
			return &types.UpdateUserProfileInfoResp{
				Code: errorcode.ErrInternalError,
				Msg:  errorcode.GetCodeMessage(errorcode.ErrInternalError),
			}, nil
		default:
			return &types.UpdateUserProfileInfoResp{
				Code: errorcode.ErrInternalError,
				Msg:  errorcode.GetCodeMessage(errorcode.ErrInternalError),
			}, nil
		}
	}

	// 5. 将 RPC 返回的 pb 对象转换为 API 响应类型
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
		Code:        errorcode.Success,
		Msg:         errorcode.GetCodeMessage(errorcode.Success),
		UserProfile: apiProfile,
	}, nil
}
