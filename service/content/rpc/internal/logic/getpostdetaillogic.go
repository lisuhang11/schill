package logic

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"SChill/common/cacheprotect"
	"SChill/common/cacheutil"
	errutil "SChill/common/error"
	"SChill/common/redis"
	"SChill/service/content/rpc/internal/model"
	"SChill/service/content/rpc/internal/svc"
	"SChill/service/content/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

var postDetailGroup cacheprotect.Group

const (
	// PostBaseTTL: immutable fields — title, cover, content, topics, created_at, user_id, visibility.
	// Long TTL with jitter. Invalidated only on post edit/delete.
	postBaseLogicalTTL  = 600 * time.Second // 10 min
	postBasePhysicalTTL = 1200 * time.Second // 20 min

	// PostDetail empty/null markers use a short TTL.
	postDetailEmptyTTL = time.Duration(redis.CacheNullExpire) * time.Second

	// Lock/wait for cache rebuild.
	postDetailLockTTL      = 10 * time.Second
	postDetailWaitInterval = 50 * time.Millisecond
	postDetailWaitAttempts = 20

	// Local cache TTL for post detail.
	localCacheTTL = time.Minute
)

type GetPostDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetPostDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPostDetailLogic {
	return &GetPostDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetPostDetailLogic) GetPostDetail(in *pb.GetPostDetailReq) (*pb.GetPostDetailResp, error) {
	if in.PostId == 0 {
		return nil, errutil.RpcBusinessError(errutil.ErrInvalidParams)
	}

	// 1. Try local cache first (hottest path)
	baseCacheKey := buildPostBaseCacheKey(in.PostId)
	if cached, ok := loadLocalCache[*pb.GetPostDetailResp](l.svcCtx.LocalCache, baseCacheKey); ok && cached != nil {
		return cached, nil
	}

	// 2. Try Redis base cache (cacheprotect with two-TTL strategy)
	cached, entry, err := l.loadCachedPostDetail(baseCacheKey)
	if err == nil && entry != nil {
		switch {
		case entry.Empty && entry.IsFresh(time.Now()):
			return nil, errutil.RpcBusinessError(errutil.ErrPostNotExist)
		case cached != nil && entry.IsFresh(time.Now()):
			storeLocalCache(l.svcCtx, baseCacheKey, cached, localCacheTTL)
			return cached, nil
		case cached != nil:
			storeLocalCache(l.svcCtx, baseCacheKey, cached, 30*time.Second)
			l.refreshPostDetailAsync(in)
			return cached, nil
		}
	}

	// 3. Cache miss → protected rebuild
	return l.loadPostDetailProtected(in, baseCacheKey)
}

func (l *GetPostDetailLogic) loadCachedPostDetail(cacheKey string) (*pb.GetPostDetailResp, *cacheprotect.Entry, error) {
	entry, err := cacheprotect.LoadEntry(l.ctx, l.svcCtx.Redis, cacheKey)
	if err != nil {
		return nil, nil, err
	}

	if entry.Empty {
		return nil, entry, nil
	}

	var cached pb.GetPostDetailResp
	if err := entry.Decode(&cached); err != nil {
		return nil, nil, err
	}
	if cached.Post == nil {
		return nil, entry, nil
	}

	return &cached, entry, nil
}

func (l *GetPostDetailLogic) loadPostDetailProtected(in *pb.GetPostDetailReq, cacheKey string) (*pb.GetPostDetailResp, error) {
	val, err, _ := postDetailGroup.Do(cacheKey, func() (interface{}, error) {
		if cached, entry, loadErr := l.loadCachedPostDetail(cacheKey); loadErr == nil && entry != nil {
			switch {
			case entry.Empty && entry.IsFresh(time.Now()):
				return nil, errutil.RpcBusinessError(errutil.ErrPostNotExist)
			case cached != nil && entry.IsFresh(time.Now()):
				return cached, nil
			}
		}

		return l.refreshPostDetailSync(in, cacheKey)
	})
	if err != nil {
		return nil, err
	}

	resp, _ := val.(*pb.GetPostDetailResp)
	if resp == nil {
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}
	return resp, nil
}

func (l *GetPostDetailLogic) refreshPostDetailAsync(in *pb.GetPostDetailReq) {
	go func(postID uint64) {
		bgLogic := NewGetPostDetailLogic(context.Background(), l.svcCtx)
		cacheKey := buildPostBaseCacheKey(postID)
		if _, err := bgLogic.loadPostDetailProtected(&pb.GetPostDetailReq{PostId: postID}, cacheKey); err != nil {
			logx.Errorf("async refresh post detail failed: postId=%d err=%v", postID, err)
		}
	}(in.PostId)
}

func (l *GetPostDetailLogic) refreshPostDetailSync(in *pb.GetPostDetailReq, cacheKey string) (*pb.GetPostDetailResp, error) {
	lockKey := fmt.Sprintf("%s%d", redis.PostDetailLockKey, in.PostId)
	acquired, err := cacheprotect.TryLock(l.ctx, l.svcCtx.Redis, lockKey, postDetailLockTTL)
	if err != nil {
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	if !acquired {
		ok, waitErr := cacheprotect.WaitFor(l.ctx, postDetailWaitAttempts, postDetailWaitInterval, func() (bool, error) {
			cached, entry, loadErr := l.loadCachedPostDetail(cacheKey)
			if loadErr != nil || entry == nil {
				return false, nil
			}
			if entry.Empty && entry.IsFresh(time.Now()) {
				return false, errutil.RpcBusinessError(errutil.ErrPostNotExist)
			}
			return cached != nil && entry.IsFresh(time.Now()), nil
		})
		if waitErr != nil {
			return nil, waitErr
		}
		if ok {
			cached, _, loadErr := l.loadCachedPostDetail(cacheKey)
			if loadErr == nil && cached != nil {
				return cached, nil
			}
		}
	}

	if acquired {
		defer func() {
			if err := cacheprotect.ReleaseLock(context.Background(), l.svcCtx.Redis, lockKey); err != nil {
				logx.Errorf("release post detail lock failed: postId=%d err=%v", in.PostId, err)
			}
		}()
	}

	return l.queryPostDetailAndCache(in, cacheKey)
}

func (l *GetPostDetailLogic) queryPostDetailAndCache(in *pb.GetPostDetailReq, cacheKey string) (*pb.GetPostDetailResp, error) {
	var post model.Post
	if err := l.svcCtx.DBRead.WithContext(l.ctx).Where("id = ? AND deleted_at IS NULL", in.PostId).First(&post).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			if cacheErr := cacheprotect.StoreEmpty(l.ctx, l.svcCtx.Redis, cacheKey, postDetailEmptyTTL, postDetailEmptyTTL); cacheErr != nil {
				logx.Errorf("cache empty post detail failed: postId=%d err=%v", in.PostId, cacheErr)
			}
			return nil, errutil.RpcBusinessError(errutil.ErrPostNotExist)
		}
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	var contents []model.PostContent
	if err := l.svcCtx.DBRead.WithContext(l.ctx).
		Where("post_id = ? AND deleted_at IS NULL", in.PostId).
		Order("sort ASC, id ASC").
		Find(&contents).Error; err != nil {
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	var postTopics []model.PostTopic
	if err := l.svcCtx.DBRead.WithContext(l.ctx).Where("post_id = ?", in.PostId).Find(&postTopics).Error; err != nil {
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	topicNameMap := make(map[uint64]string, len(postTopics))
	if len(postTopics) > 0 {
		topicIDs := make([]uint64, 0, len(postTopics))
		for _, pt := range postTopics {
			topicIDs = append(topicIDs, pt.TopicID)
		}
		var topics []model.Topic
		if err := l.svcCtx.DBRead.WithContext(l.ctx).Where("id IN ? AND deleted_at IS NULL", topicIDs).Find(&topics).Error; err != nil {
			return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
		}
		for _, topic := range topics {
			topicNameMap[topic.ID] = topic.Name
		}
	}

	resp := &pb.GetPostDetailResp{
		Post: &pb.PostInfo{
			Id:              post.ID,
			UserId:          post.UserID,
			Title:           post.Title,
			Cover:           post.Cover,
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
		},
		Contents: make([]*pb.PostContentItem, 0, len(contents)),
		Topics:   make([]*pb.PostTopic, 0, len(postTopics)),
	}

	for _, content := range contents {
		resp.Contents = append(resp.Contents, &pb.PostContentItem{
			Type:    content.Type,
			Content: content.Content,
			Sort:    content.Sort,
		})
	}
	if resp.Post.Title == "" {
		resp.Post.Title = deriveTitleFromPostContents(contents)
	}
	if resp.Post.Cover == "" {
		resp.Post.Cover = deriveCoverFromPostContents(contents)
	}
	for _, postTopic := range postTopics {
		resp.Topics = append(resp.Topics, &pb.PostTopic{
			PostId:    postTopic.PostID,
			TopicId:   postTopic.TopicID,
			TopicName: topicNameMap[postTopic.TopicID],
		})
	}

	// Store base cache with jitter TTL
	logicalTTL := cacheutil.JitterDefault(postBaseLogicalTTL)
	physicalTTL := cacheutil.JitterDefault(postBasePhysicalTTL)
	if err := cacheprotect.StoreValue(l.ctx, l.svcCtx.Redis, cacheKey, resp, logicalTTL, physicalTTL); err != nil {
		logx.Errorf("cache post detail failed: postId=%d err=%v", in.PostId, err)
	}
	if len(resp.Contents) > 0 {
		_ = l.svcCtx.Redis.SetJSON(l.ctx, buildPostContentCacheKey(l.ctx, l.svcCtx, in.PostId), resp.Contents, physicalTTL)
	}
	storeLocalCache(l.svcCtx, cacheKey, resp, localCacheTTL)

	return resp, nil
}

// buildPostBaseCacheKey constructs the base cache key (immutable fields).
// The version is embedded for cache busting on post edit/delete.
func buildPostBaseCacheKey(postID uint64) string {
	return fmt.Sprintf("%s%d", redis.PostBaseKey, postID)
}

// buildPostStatsCacheKey constructs the stats cache key (counters, short TTL).
func buildPostStatsCacheKey(postID uint64) string {
	return fmt.Sprintf("%s%s", redis.PostStatsKey, strconv.FormatUint(postID, 10))
}
