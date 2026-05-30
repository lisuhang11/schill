package logic

import (
	"context"
	"strings"
	"time"

	"SChill/common/redis"
	"SChill/service/content/rpc/internal/model"
	"SChill/service/content/rpc/internal/svc"
	"SChill/service/content/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/metadata"
)

type GetPostListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetPostListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPostListLogic {
	return &GetPostListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetPostListLogic) GetPostList(in *pb.GetPostListReq) (*pb.GetPostListResp, error) {
	page := in.Page
	if page <= 0 {
		page = 1
	}
	pageSize := in.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	feedType := feedTypeFromContext(l.ctx)
	useCache := feedType != feedTypeFollowing
	cacheKey := ""
	if useCache {
		cacheKey = buildPostListCacheKey(l.ctx, l.svcCtx, in.UserId, feedType, page, pageSize)
		if cached, ok := loadLocalCache[*pb.GetPostListResp](l.svcCtx.LocalCache, cacheKey); ok && cached != nil {
			return cached, nil
		}

		var cached pb.GetPostListResp
		if err := l.svcCtx.Cache.GetCtx(l.ctx, cacheKey, &cached); err == nil {
			storeLocalCache(l.svcCtx, cacheKey, &cached, time.Minute)
			return &cached, nil
		}
	}

	query := l.svcCtx.DBRead.WithContext(l.ctx).Model(&model.Post{}).Where("post.deleted_at IS NULL")
	if in.UserId > 0 {
		query = query.Where("post.user_id = ?", in.UserId)
	}
	if feedType == feedTypeFollowing {
		query = query.Joins("JOIN following ON following.follow_id = post.user_id").
			Where("following.user_id = ?", in.CurrentUserId)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var posts []model.Post
	if err := query.
		Order("post.is_top DESC, post.created_at DESC").
		Offset(int((page - 1) * pageSize)).
		Limit(int(pageSize)).
		Find(&posts).Error; err != nil {
		return nil, err
	}

	resp := &pb.GetPostListResp{
		Total: total,
		List:  make([]*pb.PostInfo, 0, len(posts)),
	}
	derivedMeta := l.batchLoadDerivedPostMeta(posts)
	for _, post := range posts {
		title := post.Title
		cover := post.Cover
		if meta, ok := derivedMeta[post.ID]; ok {
			if title == "" {
				title = meta.Title
			}
			if cover == "" {
				cover = meta.Cover
			}
		}
		resp.List = append(resp.List, &pb.PostInfo{
			Id:              post.ID,
			UserId:          post.UserID,
			Title:           title,
			Cover:           cover,
			CommentCount:    post.CommentCount,
			CollectionCount: post.CollectionCount,
			UpvoteCount:     post.UpvoteCount,
			ShareCount:      post.ShareCount,
			Visibility:      post.Visibility,
			IsTop:           post.IsTop,
			IsEssence:       post.IsEssence,
			IsLock:          post.IsLock,
			LatestRepliedAt: post.LatestRepliedAt,
			Tags:            post.Tags,
			CreatedAt:       post.CreatedAt.Unix(),
			UpdatedAt:       post.UpdatedAt.Unix(),
		})
	}

	if useCache {
		if err := l.svcCtx.Cache.SetWithExpireCtx(l.ctx, cacheKey, resp, time.Duration(redis.PostExpire)*time.Second); err != nil {
			logx.Errorf("cache post list failed: %v", err)
		}
		storeLocalCache(l.svcCtx, cacheKey, resp, time.Minute)
	}

	return resp, nil
}

const (
	feedTypeLatest    = "latest"
	feedTypeFollowing = "following"
)

func feedTypeFromContext(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return feedTypeLatest
	}
	values := md.Get("x-feed-type")
	if len(values) == 0 {
		return feedTypeLatest
	}
	switch strings.ToLower(strings.TrimSpace(values[0])) {
	case feedTypeFollowing:
		return feedTypeFollowing
	default:
		return feedTypeLatest
	}
}

func (l *GetPostListLogic) batchLoadDerivedPostMeta(posts []model.Post) map[uint64]postDerivedMeta {
	postIDs := make([]uint64, 0, len(posts))
	for _, post := range posts {
		if post.Title == "" || post.Cover == "" {
			postIDs = append(postIDs, post.ID)
		}
	}
	if len(postIDs) == 0 {
		return nil
	}

	var contents []model.PostContent
	if err := l.svcCtx.DBRead.WithContext(l.ctx).
		Where("post_id IN ? AND deleted_at IS NULL", postIDs).
		Order("post_id ASC, sort ASC, id ASC").
		Find(&contents).Error; err != nil {
		logx.Errorf("load post content for derived meta failed: %v", err)
		return nil
	}

	result := make(map[uint64]postDerivedMeta, len(postIDs))
	for _, content := range contents {
		meta := result[content.PostID]
		if meta.Title == "" {
			meta.Title = deriveTitleFromContentString(content.Content)
		}
		if meta.Cover == "" && content.Type == 3 {
			meta.Cover = strings.TrimSpace(content.Content)
		}
		result[content.PostID] = meta
	}
	return result
}
