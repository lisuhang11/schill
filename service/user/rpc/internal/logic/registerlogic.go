package logic

import (
	errutil "SChill/common/error"
	"context"
	"errors"
	"unicode"

	"gorm.io/gorm/clause"
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

	if err := validatePassword(password); err != nil {
		return nil, err
	}

	hashedPwd := cryptx.PasswordEncrypt(password)
	user := &model.User{
		Username:     username,
		PasswordHash: hashedPwd,
		Avatar:       "http://localhost:9000/user-avatar/user_default_avatar.png",
		Status:       1,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	var userId uint64
	err := l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		// 原子插入，如果 username 冲突则忽略插入
		result := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "username"}}, // 指定唯一键列
			DoNothing: true,
		}).Create(user)

		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			// 用户名已存在（包括软删除的记录，因为唯一索引包含删除态）
			return errutil.RpcBusinessError(errutil.ErrUsernameExists)
		}
		userId = user.ID

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
		if errors.Is(err, errutil.RpcBusinessError(errutil.ErrUsernameExists)) {
			return nil, err
		}
		logx.Errorf("注册事务失败: %v", err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	return &pb.RegisterResp{UserId: userId}, nil
}

// validatePassword checks password strength:
// - minimum 8 characters
// - at least one uppercase letter
// - at least one lowercase letter
// - at least one digit
func validatePassword(password string) error {
	if len(password) < 8 {
		return errutil.RpcBusinessError(errutil.ErrPasswordTooWeak)
	}

	var hasUpper, hasLower, hasDigit bool
	for _, ch := range password {
		switch {
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsLower(ch):
			hasLower = true
		case unicode.IsDigit(ch):
			hasDigit = true
		}
	}

	if !hasUpper || !hasLower || !hasDigit {
		return errutil.RpcBusinessError(errutil.ErrPasswordTooWeak)
	}

	return nil
}
