package logic

import (
	errutil "SChill/common/error"
	"context"
	"time"

	"SChill/common/cryptx"
	"SChill/service/user/rpc/internal/model"
	"SChill/service/user/rpc/internal/svc"
	"SChill/service/user/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type RegisterLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RegisterLogic) Register(in *pb.RegisterReq) (*pb.RegisterResp, error) {
	username := in.Username
	password := in.Password

	// 1. 检查用户名是否已存在（包括软删除）
	var existingUser model.User
	err := l.svcCtx.DB.WithContext(l.ctx).Where("username = ?", username).First(&existingUser).Error
	if err == nil {
		return nil, errutil.RpcBusinessError(errutil.ErrUsernameExists)
	}
	if err != gorm.ErrRecordNotFound {
		logx.Errorf("查询用户失败: %v", err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	// 2. 准备用户数据
	hashedPwd := cryptx.PasswordEncrypt(password)
	user := &model.User{
		Username:     username,
		PasswordHash: hashedPwd,
		Status:       1,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// 3. 开启事务
	err = l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		// 插入 user
		if err := tx.Create(user).Error; err != nil {
			return err
		}

		// 插入 user_profile
		profile := &model.UserProfile{
			UserID:    user.ID,
			Gender:    0,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := tx.Create(profile).Error; err != nil {
			return err
		}

		// 插入 user_stat
		stat := &model.UserStat{
			UserID:          user.ID,
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
		if err := tx.Create(stat).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		logx.Errorf("注册用户的数据库事务失败: %v", err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	return &pb.RegisterResp{
		UserId: user.ID,
	}, nil
}
