package logic

import (
	"SChill/common/errorcode"
	"SChill/service/user/api/internal/svc"
	"SChill/service/user/api/internal/types"
	"SChill/service/user/rpc/pb"
	"context"
	"github.com/zeromicro/go-zero/core/logc"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RegisterLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegisterLogic) Register(req *types.RegisterReq) (resp *types.RegisterResp, err error) {
	// 1. 参数校验
	if err := req.Validate(); err != nil {
		logc.Infof(l.ctx, "注册参数错误: %v", err)
		return &types.RegisterResp{
			Code: errorcode.ErrUsernameOrPasswordEmpty, // 用户名和密码参数
			Msg:  errorcode.GetCodeMessage(errorcode.ErrUsernameOrPasswordEmpty),
		}, nil
	}

	// 2. 调用 RPC 注册服务
	rpcResp, err := l.svcCtx.UserRpc.Register(l.ctx, &pb.RegisterReq{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		logc.Errorf(l.ctx, "注册 RPC 调用失败: %v", err)

		// 解析 gRPC 错误状态，映射为业务错误码
		st, _ := status.FromError(err)
		switch st.Code() {
		case codes.AlreadyExists:
			// 用户名已存在
			return &types.RegisterResp{
				Code: errorcode.ErrUsernameExists,
				Msg:  errorcode.GetCodeMessage(errorcode.ErrUsernameExists),
			}, nil
		case codes.InvalidArgument:
			// 参数无效（如密码格式不对），可根据实际情况细化
			return &types.RegisterResp{
				Code: errorcode.ErrUsernameOrPasswordEmpty, // 或定义更具体的错误码
				Msg:  errorcode.GetCodeMessage(errorcode.ErrUsernameOrPasswordEmpty),
			}, nil
		default:
			// 其他内部错误
			return &types.RegisterResp{
				Code: errorcode.ErrInternalError,
				Msg:  errorcode.GetCodeMessage(errorcode.ErrInternalError),
			}, nil
		}
	}

	// 3. 成功响应
	return &types.RegisterResp{
		Code:   errorcode.Success,
		Msg:    errorcode.GetCodeMessage(errorcode.Success),
		UserId: rpcResp.UserId,
	}, nil
}
