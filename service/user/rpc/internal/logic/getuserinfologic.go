package logic

import (
	"context"
	"time"

	errutil "SChill/common/error"
	"SChill/service/user/rpc/internal/svc"
	"SChill/service/user/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type GetUserInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserInfoLogic {
	return &GetUserInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ==================== 查询结果结构体（用于 LEFT JOIN）====================
type UserJoinResult struct {
	// user 表
	UserID    uint64
	Username  string
	Phone     *string
	Email     *string
	Avatar    string
	Status    int8
	IsAdmin   int8
	CreatedAt time.Time

	// profile 表（可能为 NULL）
	ProfileUserID *uint64
	Gender        *int8
	Birthday      *time.Time
	Signature     *string
	Location      *string
	Website       *string
	Company       *string
	JobTitle      *string
	Education     *string

	// stat 表（可能为 NULL）
	StatUserID      *uint64
	PostCount       *uint32
	CommentCount    *uint32
	FollowerCount   *uint32
	FollowingCount  *uint32
	LikeCount       *uint32
	CollectionCount *uint32
	LastActiveTime  *int64 // DB 侧用 UNIX_TIMESTAMP 转换
}

// ==================== 主逻辑 ====================
func (l *GetUserInfoLogic) GetUserInfo(in *pb.GetUserInfoReq) (*pb.GetUserInfoResp, error) {
	userId := in.UserId
	cacheKey := buildUserInfoCacheKey(userId)

	var resp pb.GetUserInfoResp
	err := l.svcCtx.Cache.TakeCtx(l.ctx, &resp, cacheKey, func(val interface{}) error {
		loadCtx, cancel := context.WithTimeout(l.ctx, 3*time.Second)
		defer cancel()

		dbResp, err := l.loadUserFromDB(loadCtx, userId)
		if err != nil {
			return err
		}

		*val.(*pb.GetUserInfoResp) = *dbResp
		return nil
	})

	if err != nil {
		return nil, err
	}

	if resp.UserInfo == nil {
		return nil, errutil.RpcBusinessError(errutil.ErrUserNotExist)
	}

	return &resp, nil
}

// loadUserFromDB 使用 LEFT JOIN 一次查询三张表
// ctx 由调用方控制超时和取消隔离
func (l *GetUserInfoLogic) loadUserFromDB(ctx context.Context, userId uint64) (*pb.GetUserInfoResp, error) {
	var result UserJoinResult
	err := l.svcCtx.DB.WithContext(ctx).
		Table("user").
		Select(`
			user.user_id,
			user.username,
			user.phone,
			user.email,
			user.avatar,
			user.status,
			user.is_admin,
			user.created_at,
			profile.user_id AS profile_user_id,
			profile.gender,
			profile.birthday,
			profile.signature,
			profile.location,
			profile.website,
			profile.company,
			profile.job_title,
			profile.education,
			stat.user_id AS stat_user_id,
			stat.post_count,
			stat.comment_count,
			stat.follower_count,
			stat.following_count,
			stat.like_count,
			stat.collection_count,
			CAST(FLOOR(UNIX_TIMESTAMP(stat.last_active_time)) AS SIGNED) AS last_active_time
		`).
		Joins("LEFT JOIN user_profile profile ON user.user_id = profile.user_id").
		Joins("LEFT JOIN user_stat stat ON user.user_id = stat.user_id").
		Where("user.user_id = ? AND user.deleted_at IS NULL", userId).
		Take(&result).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 用户不存在，返回空标记（UserInfo 为 nil）
			return &pb.GetUserInfoResp{}, nil
		}
		logx.Errorf("数据库查询失败: userId=%d, err=%v", userId, err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	// 组装响应（user 字段一定存在）
	userInfo := &pb.UserInfo{
		UserId:    result.UserID,
		Username:  result.Username,
		Phone:     getStringValue(result.Phone),
		Email:     getStringValue(result.Email),
		Avatar:    result.Avatar,
		Status:    int32(result.Status),
		IsAdmin:   result.IsAdmin == 1,
		CreatedAt: result.CreatedAt.Unix(),
	}
	resp := &pb.GetUserInfoResp{UserInfo: userInfo}

	// Profile 存在性判断
	if result.ProfileUserID != nil {
		gender := int64(getInt8Value(result.Gender))
		signature := getStringValue(result.Signature)
		location := getStringValue(result.Location)
		website := getStringValue(result.Website)
		company := getStringValue(result.Company)
		jobTitle := getStringValue(result.JobTitle)
		education := getStringValue(result.Education)
		pbProfile := &pb.UserProfile{
			UserId:    *result.ProfileUserID,
			Gender:    &gender,
			Signature: &signature,
			Location:  &location,
			Website:   &website,
			Company:   &company,
			JobTitle:  &jobTitle,
			Education: &education,
		}
		if result.Birthday != nil {
			birthday := result.Birthday.Format("2006-01-02")
			pbProfile.Birthday = &birthday
		}
		resp.Profile = pbProfile
	}

	// Stat 存在性判断
	if result.StatUserID != nil {
		resp.Stat = &pb.UserStat{
			UserId:          *result.StatUserID,
			PostCount:       uint64(getUint32Value(result.PostCount)),
			CommentCount:    uint64(getUint32Value(result.CommentCount)),
			FollowerCount:   uint64(getUint32Value(result.FollowerCount)),
			FollowingCount:  uint64(getUint32Value(result.FollowingCount)),
			LikeCount:       uint64(getUint32Value(result.LikeCount)),
			CollectionCount: uint64(getUint32Value(result.CollectionCount)),
			LastActiveTime:  getInt64Value(result.LastActiveTime),
		}
	}

	return resp, nil
}

// ==================== 辅助函数 ====================

func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func getInt8Value(i *int8) int8 {
	if i == nil {
		return 0
	}
	return *i
}

func getUint32Value(u *uint32) uint32 {
	if u == nil {
		return 0
	}
	return *u
}

func getInt64Value(i *int64) int64 {
	if i == nil {
		return 0
	}
	return *i
}
