package logic

import (
	"context"
	"database/sql"
	"github.com/zeromicro/go-zero/core/logc"
	"time"

	"SChill/common/cryptx"
	"SChill/service/user/rpc/internal/model"
	"SChill/service/user/rpc/internal/svc"
	"SChill/service/user/rpc/pb"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
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

func (l *RegisterLogic) Register(in *pb.RegisterReq) (*pb.RegisterResp, error) {
	username := in.Username
	password := in.Password

	// 1. 检查用户名是否已存在
	_, err := l.svcCtx.UserModel.FindOneByUsername(l.ctx, username)
	switch {
	case err == nil:
		return nil, status.Error(codes.AlreadyExists, "用户名已存在")
	case err != model.ErrNotFound:
		logc.Errorf(l.ctx, "查询用户失败: %v", err)
		return nil, status.Error(codes.Internal, "内部错误")
	}

	// 2. 准备用户数据
	hashedPwd := cryptx.PasswordEncrypt(password)
	user := &model.User{
		Username:      username,
		PasswordHash:  hashedPwd,
		Status:        1, // 正常状态
		LastLoginTime: sql.NullTime{Valid: false},
		DeletedAt:     sql.NullTime{Valid: false},
	}

	// 3. 开启事务，插入三张表
	var userId uint64
	err = l.svcCtx.Conn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		// 从 session 创建事务专用连接
		txConn := sqlx.NewSqlConnFromSession(session)

		// 重新实例化带事务的 Model
		userModel := model.NewUserModel(txConn)
		profileModel := model.NewUserProfileModel(txConn)
		statModel := model.NewUserStatModel(txConn)

		// 3.1 插入 user 表
		result, err := userModel.Insert(ctx, user)
		if err != nil {
			return err
		}
		id, err := result.LastInsertId()
		if err != nil {
			return err
		}
		userId = uint64(id)

		// 3.2 插入 user_profile
		profile := &model.UserProfile{
			UserId:    userId,
			Gender:    0,
			Birthday:  sql.NullTime{Valid: false},
			Signature: "",
			Location:  "",
			Website:   "",
			Company:   "",
			JobTitle:  "",
			Education: "",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		_, err = profileModel.Insert(ctx, profile)
		if err != nil {
			return err
		}

		// 3.3 插入 user_stat
		stat := &model.UserStat{
			UserId:          userId,
			PostCount:       0,
			CommentCount:    0,
			FollowerCount:   0,
			FollowingCount:  0,
			LikeCount:       0,
			CollectionCount: 0,
			LastActiveTime:  0,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}
		_, err = statModel.Insert(ctx, stat)
		return err
	})

	if err != nil {
		logc.Errorf(l.ctx, "注册用户的数据库事务失败: %v", err)
		return nil, status.Error(codes.Internal, "注册失败")
	}

	return &pb.RegisterResp{
		UserId: userId,
	}, nil
}
