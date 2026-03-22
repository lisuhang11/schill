package logic

import (
	"context"
	"database/sql"
	"time"

	"SChill/common/cryptx"
	"SChill/service/user/rpc/internal/model"
	"SChill/service/user/rpc/internal/svc"
	"SChill/service/user/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *LoginLogic) Login(in *pb.LoginReq) (*pb.LoginResp, error) {
	// 1. 根据用户名查找用户
	user, err := l.svcCtx.UserModel.FindOneByUsername(l.ctx, in.Username)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, status.Error(codes.NotFound, "用户不存在")
		}
		logx.Errorf("查询用户失败: %v", err)
		return nil, status.Error(codes.Internal, "内部错误")
	}

	// 2. 检查软删除
	if user.DeletedAt.Valid {
		return nil, status.Error(codes.NotFound, "用户不存在")
	}

	// 3. 验证密码
	if !cryptx.PasswordVerify(user.PasswordHash, in.Password) {
		return nil, status.Error(codes.Unauthenticated, "密码错误")
	}

	// 4. 检查用户状态（1正常，2禁言，3冻结）
	if user.Status != 1 {
		return nil, status.Error(codes.PermissionDenied, "账号异常，无法登录")
	}

	// 5. 更新最后登录时间（IP 未记录，可按需从请求元数据中获取）
	user.LastLoginTime = sql.NullTime{
		Time:  time.Now(),
		Valid: true,
	}
	if err := l.svcCtx.UserModel.Update(l.ctx, user); err != nil {
		logx.Errorf("更新用户登录时间失败: %v", err)
		// 登录成功但更新失败不影响主流程
	}

	// 6. 返回用户 ID
	return &pb.LoginResp{
		UserId: user.Id,
	}, nil
}
