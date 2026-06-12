package logic

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"SChill/common/cacheprotect"
	"SChill/common/commentrank"
	errutil "SChill/common/error"
	"SChill/common/redis"
	"SChill/service/comment/rpc/internal/model"
	"SChill/service/comment/rpc/internal/svc"
	"SChill/service/comment/rpc/pb"

	goredis "github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
)

var commentListGroup cacheprotect.Group

const (
	commentListLogicalTTL   = time.Duration(redis.CommentExpire) * time.Second
	commentListPhysicalTTL  = commentListLogicalTTL * 2
	commentListEmptyTTL     = time.Duration(redis.CacheNullExpire) * time.Second
	commentListLockTTL      = 10 * time.Second
	commentListWaitInterval = 50 * time.Millisecond
	commentListWaitAttempts = 20
)

type commentListCacheState struct {
	IDs        []uint64
	HasMore    bool
	NextCursor int64
	Entry      *cacheprotect.Entry
}

type GetCommentListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCommentListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCommentListLogic {
	return &GetCommentListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetCommentListLogic) GetCommentList(in *pb.GetCommentListReq) (*pb.GetCommentListResp, error) {
	if in.PostId == 0 {
		return nil, errutil.RpcBusinessError(errutil.ErrInvalidParams)
	}

	pageSize := in.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	cacheState, err := l.getRootCommentIDsFromRedis(in.PostId, in.Cursor, pageSize)
	if err != nil {
		logx.Errorf("get comment ids from redis failed: %v", err)
		return l.getCommentListFromDB(in, pageSize)
	}

	if cacheState.Entry != nil && cacheState.Entry.Empty && cacheState.Entry.IsFresh(time.Now()) {
		return &pb.GetCommentListResp{
			Total:      0,
			List:       []*pb.CommentInfo{},
			HasMore:    false,
			NextCursor: 0,
		}, nil
	}

	if len(cacheState.IDs) == 0 {
		if err := l.ensureCommentCache(in.PostId); err != nil {
			logx.Errorf("ensure comment cache failed: %v", err)
			return l.getCommentListFromDB(in, pageSize)
		}
		cacheState, err = l.getRootCommentIDsFromRedis(in.PostId, in.Cursor, pageSize)
		if err != nil {
			logx.Errorf("get rebuilt comment ids from redis failed: %v", err)
			return l.getCommentListFromDB(in, pageSize)
		}
		if cacheState.Entry != nil && cacheState.Entry.Empty && cacheState.Entry.IsFresh(time.Now()) {
			return &pb.GetCommentListResp{
				Total:      0,
				List:       []*pb.CommentInfo{},
				HasMore:    false,
				NextCursor: 0,
			}, nil
		}
	} else if cacheState.Entry == nil || !cacheState.Entry.IsFresh(time.Now()) {
		l.refreshCommentCacheAsync(in.PostId)
	}

	commentInfoMap, err := l.batchGetCommentInfo(cacheState.IDs)
	if err != nil {
		logx.Errorf("batch get comment info failed: %v", err)
		return l.getCommentListFromDB(in, pageSize)
	}

	contentMap, err := l.batchGetCommentContent(cacheState.IDs)
	if err != nil {
		logx.Errorf("batch get comment content failed: %v", err)
	}

	list := l.assembleCommentList(cacheState.IDs, commentInfoMap, contentMap)
	total := int64(0)
	if !(cacheState.Entry != nil && cacheState.Entry.Empty && len(cacheState.IDs) == 0) {
		total, _ = l.getTotalCommentCount(in.PostId)
	}

	return &pb.GetCommentListResp{
		Total:      total,
		List:       list,
		HasMore:    cacheState.HasMore,
		NextCursor: cacheState.NextCursor,
	}, nil
}

func (l *GetCommentListLogic) getRootCommentIDsFromRedis(postId uint64, cursor, pageSize int64) (*commentListCacheState, error) {
	ctx := context.Background()
	entry, err := cacheprotect.LoadEntry(ctx, l.svcCtx.Redis, buildCommentListMetaKey(postId))
	if err != nil {
		entry = nil
	}

	key := buildCommentListKey(postId)

	maxScore := "+inf"
	minScore := "-inf"
	if cursor > 0 {
		maxScore = fmt.Sprintf("(%d", cursor)
	}

	// Use ZRevRangeByScoreWithScores to get both member and score
	zs, err := l.svcCtx.Redis.ZRevRangeByScoreWithScores(ctx, key, &goredis.ZRangeBy{
		Min:    minScore,
		Max:    maxScore,
		Offset: 0,
		Count:  pageSize + 1,
	}).Result()
	if err != nil {
		return nil, err
	}

	hasMore := len(zs) > int(pageSize)
	if hasMore {
		zs = zs[:pageSize]
	}

	ids := make([]uint64, 0, len(zs))
	for _, z := range zs {
		id, _ := strconv.ParseUint(z.Member.(string), 10, 64)
		ids = append(ids, id)
	}

	var nextCursor int64
	if len(zs) > 0 {
		// Use the score of the last item as the cursor for the next page
		nextCursor = int64(zs[len(zs)-1].Score)
	}

	return &commentListCacheState{
		IDs:        ids,
		HasMore:    hasMore,
		NextCursor: nextCursor,
		Entry:      entry,
	}, nil
}

func (l *GetCommentListLogic) ensureCommentCache(postId uint64) error {
	flightKey := fmt.Sprintf("comment:list:%d", postId)
	_, err, _ := commentListGroup.Do(flightKey, func() (interface{}, error) {
		cacheState, loadErr := l.getRootCommentIDsFromRedis(postId, 0, 1)
		if loadErr == nil && cacheState != nil && cacheState.Entry != nil && cacheState.Entry.IsFresh(time.Now()) {
			return nil, nil
		}
		return nil, l.rebuildCommentCache(postId)
	})
	return err
}

func (l *GetCommentListLogic) refreshCommentCacheAsync(postId uint64) {
	go func() {
		bgLogic := NewGetCommentListLogic(context.Background(), l.svcCtx)
		if err := bgLogic.ensureCommentCache(postId); err != nil {
			logx.Errorf("async refresh comment cache failed: postId=%d err=%v", postId, err)
		}
	}()
}

func (l *GetCommentListLogic) rebuildCommentCache(postId uint64) error {
	ctx := context.Background()
	lockKey := buildCommentListLockKey(postId)

	acquired, err := cacheprotect.TryLock(ctx, l.svcCtx.Redis, lockKey, commentListLockTTL)
	if err != nil {
		return err
	}
	if !acquired {
		ok, waitErr := cacheprotect.WaitFor(ctx, commentListWaitAttempts, commentListWaitInterval, func() (bool, error) {
			cacheState, loadErr := l.getRootCommentIDsFromRedis(postId, 0, 1)
			if loadErr != nil || cacheState == nil || cacheState.Entry == nil {
				return false, nil
			}
			return cacheState.Entry.IsFresh(time.Now()), nil
		})
		if waitErr != nil {
			return waitErr
		}
		if ok {
			return nil
		}
		return fmt.Errorf("comment cache rebuild lock not acquired")
	}
	defer cacheprotect.ReleaseLock(context.Background(), l.svcCtx.Redis, lockKey)

	var comments []*model.Comment
	query := l.svcCtx.DB.WithContext(ctx).Where("post_id = ? AND parent_id = 0 AND deleted_at IS NULL", postId)
	if err := query.Find(&comments).Error; err != nil {
		return err
	}

	key := buildCommentListKey(postId)
	metaKey := buildCommentListMetaKey(postId)
	_ = l.svcCtx.Redis.Del(ctx, key)

	members := make([]redis.Z, 0, len(comments))
	for _, c := range comments {
		score := commentrank.Score(int64(c.LikeCount), int64(c.DislikeCount), int64(c.ReplyCount), c.CreatedAt)
		members = append(members, redis.Z{Score: score, Member: c.ID})
	}
	if len(members) > 0 {
		if err := l.svcCtx.Redis.ZAdd(ctx, key, members...); err != nil {
			return err
		}
		_, _ = l.svcCtx.Redis.Expire(ctx, key, commentListPhysicalTTL)
		if err := cacheprotect.StoreMarker(ctx, l.svcCtx.Redis, metaKey, commentListLogicalTTL, commentListPhysicalTTL); err != nil {
			return err
		}
	} else {
		if err := cacheprotect.StoreEmpty(ctx, l.svcCtx.Redis, metaKey, commentListEmptyTTL, commentListEmptyTTL); err != nil {
			return err
		}
		_ = l.svcCtx.Redis.Set(ctx, fmt.Sprintf("%s%d", redis.PostCommentCountKey, postId), "0", commentListEmptyTTL)
	}

	return nil
}

func (l *GetCommentListLogic) batchGetCommentInfo(commentIDs []uint64) (map[uint64]map[string]string, error) {
	ctx := context.Background()
	result := make(map[uint64]map[string]string, len(commentIDs))

	pipe := l.svcCtx.Redis.Pipeline()
	cmds := make(map[uint64]*goredis.MapStringStringCmd, len(commentIDs))
	for _, id := range commentIDs {
		cmds[id] = pipe.HGetAll(ctx, fmt.Sprintf("%s%d", redis.CommentInfoKey, id))
	}
	if _, err := pipe.Exec(ctx); err != nil && err != goredis.Nil {
		logx.Errorf("pipeline comment info exec failed: %v", err)
	}
	for id, cmd := range cmds {
		info, err := cmd.Result()
		if err != nil && err != goredis.Nil {
			continue
		}
		if len(info) > 0 {
			result[id] = info
		}
	}

	if len(result) == len(commentIDs) {
		return result, nil
	}

	var dbComments []*model.Comment
	if err := l.svcCtx.DB.WithContext(ctx).Where("id IN ?", commentIDs).Find(&dbComments).Error; err == nil {
		for _, c := range dbComments {
			if _, ok := result[c.ID]; ok {
				continue
			}

			info := map[string]interface{}{
				"id":            c.ID,
				"post_id":       c.PostID,
				"user_id":       c.UserID,
				"parent_id":     c.ParentID,
				"level":         c.Level,
				"reply_count":   c.ReplyCount,
				"like_count":    c.LikeCount,
				"dislike_count": c.DislikeCount,
				"created_at":    c.CreatedAt.Unix(),
			}
			if c.ReplyToUserID != nil {
				info["reply_to_user_id"] = *c.ReplyToUserID
			}
			key := fmt.Sprintf("%s%d", redis.CommentInfoKey, c.ID)
			_ = l.svcCtx.Redis.HMSet(ctx, key, info)
			_, _ = l.svcCtx.Redis.Expire(ctx, key, time.Duration(redis.CommentExpire)*time.Second)

			item := make(map[string]string, len(info))
			for k, v := range info {
				item[k] = fmt.Sprintf("%v", v)
			}
			result[c.ID] = item
		}
	}

	return result, nil
}

func (l *GetCommentListLogic) batchGetCommentContent(commentIDs []uint64) (map[uint64]string, error) {
	ctx := context.Background()
	result := make(map[uint64]string, len(commentIDs))

	pipe := l.svcCtx.Redis.Pipeline()
	cmds := make(map[uint64]*goredis.StringCmd, len(commentIDs))
	for _, id := range commentIDs {
		cmds[id] = pipe.Get(ctx, fmt.Sprintf("%s%d", redis.CommentContentKey, id))
	}
	if _, err := pipe.Exec(ctx); err != nil && err != goredis.Nil {
		logx.Errorf("pipeline comment content exec failed: %v", err)
	}
	for id, cmd := range cmds {
		content, err := cmd.Result()
		if err == nil && content != "" {
			result[id] = content
		}
	}

	if len(result) == len(commentIDs) {
		return result, nil
	}

	var dbContents []*model.CommentContent
	if err := l.svcCtx.DB.WithContext(ctx).Where("comment_id IN ?", commentIDs).Find(&dbContents).Error; err == nil {
		for _, cc := range dbContents {
			if _, ok := result[cc.CommentID]; ok {
				continue
			}
			key := fmt.Sprintf("%s%d", redis.CommentContentKey, cc.CommentID)
			_ = l.svcCtx.Redis.Set(ctx, key, cc.Content, time.Duration(redis.CommentExpire)*time.Second)
			result[cc.CommentID] = cc.Content
		}
	}

	return result, nil
}

func (l *GetCommentListLogic) mapToCommentInfo(id uint64, info map[string]string, content string) *pb.CommentInfo {
	if len(info) == 0 {
		return nil
	}

	comment := &pb.CommentInfo{Id: id, Content: content}
	if val, ok := info["post_id"]; ok {
		comment.PostId, _ = strconv.ParseUint(val, 10, 64)
	}
	if val, ok := info["user_id"]; ok {
		comment.UserId, _ = strconv.ParseUint(val, 10, 64)
	}
	if val, ok := info["parent_id"]; ok {
		comment.ParentId, _ = strconv.ParseUint(val, 10, 64)
	}
	if val, ok := info["reply_to_user_id"]; ok {
		comment.ReplyToUserId, _ = strconv.ParseUint(val, 10, 64)
	}
	if val, ok := info["level"]; ok {
		level, _ := strconv.Atoi(val)
		comment.Level = int32(level)
	}
	if val, ok := info["reply_count"]; ok {
		count, _ := strconv.Atoi(val)
		comment.ReplyCount = int32(count)
	}
	if val, ok := info["like_count"]; ok {
		count, _ := strconv.Atoi(val)
		comment.LikeCount = int32(count)
	}
	if val, ok := info["dislike_count"]; ok {
		count, _ := strconv.Atoi(val)
		comment.DislikeCount = int32(count)
	}
	if val, ok := info["created_at"]; ok {
		ts, _ := strconv.ParseInt(val, 10, 64)
		comment.CreatedAt = ts
	}

	return comment
}

func (l *GetCommentListLogic) assembleCommentList(rootIDs []uint64, infoMap map[uint64]map[string]string, contentMap map[uint64]string) []*pb.CommentInfo {
	list := make([]*pb.CommentInfo, 0, len(rootIDs))
	for _, id := range rootIDs {
		if info, ok := infoMap[id]; ok {
			if comment := l.mapToCommentInfo(id, info, contentMap[id]); comment != nil {
				list = append(list, comment)
			}
		}
	}
	return list
}

func (l *GetCommentListLogic) getCommentListFromDB(in *pb.GetCommentListReq, pageSize int64) (*pb.GetCommentListResp, error) {
	var comments []*model.Comment
	query := l.svcCtx.DB.WithContext(l.ctx).Where("post_id = ? AND parent_id = 0 AND deleted_at IS NULL", in.PostId)
	if in.Cursor > 0 {
		query = query.Where("id < ?", in.Cursor)
	}
	query = query.Order("created_at DESC")

	if err := query.Limit(int(pageSize + 1)).Find(&comments).Error; err != nil {
		logx.Errorf("query comment list failed: %v", err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	hasMore := len(comments) > int(pageSize)
	if hasMore {
		comments = comments[:pageSize]
	}

	commentIDs := make([]uint64, 0, len(comments))
	for _, c := range comments {
		commentIDs = append(commentIDs, c.ID)
	}

	contentMap := make(map[uint64]*model.CommentContent, len(commentIDs))
	if len(commentIDs) > 0 {
		var commentContents []*model.CommentContent
		if err := l.svcCtx.DB.WithContext(l.ctx).Where("comment_id IN ?", commentIDs).Find(&commentContents).Error; err != nil {
			logx.Errorf("query comment content failed: %v", err)
			return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
		}
		for _, cc := range commentContents {
			contentMap[cc.CommentID] = cc
		}
	}

	list := make([]*pb.CommentInfo, 0, len(comments))
	for _, c := range comments {
		replyToUserID := uint64(0)
		if c.ReplyToUserID != nil {
			replyToUserID = *c.ReplyToUserID
		}
		content := ""
		if cc, ok := contentMap[c.ID]; ok {
			content = cc.Content
		}
		list = append(list, &pb.CommentInfo{
			Id:            c.ID,
			PostId:        c.PostID,
			UserId:        c.UserID,
			ParentId:      c.ParentID,
			ReplyToUserId: replyToUserID,
			Content:       content,
			Level:         int32(c.Level),
			ReplyCount:    c.ReplyCount,
			LikeCount:     c.LikeCount,
			DislikeCount:  c.DislikeCount,
			CreatedAt:     c.CreatedAt.Unix(),
		})
	}

	var nextCursor int64
	if hasMore && len(comments) > 0 {
		nextCursor = int64(comments[len(comments)-1].ID)
	}

	total, err := l.getTotalCommentCount(in.PostId)
	if err != nil {
		logx.Errorf("query comment total failed: %v", err)
	}

	return &pb.GetCommentListResp{
		Total:      total,
		List:       list,
		HasMore:    hasMore,
		NextCursor: nextCursor,
	}, nil
}

func (l *GetCommentListLogic) getTotalCommentCount(postId uint64) (int64, error) {
	cacheKey := fmt.Sprintf("%s%d", redis.PostCommentCountKey, postId)
	if val, err := l.svcCtx.Redis.Get(l.ctx, cacheKey); err == nil && val != "" {
		if cached, parseErr := strconv.ParseInt(val, 10, 64); parseErr == nil {
			return cached, nil
		}
	}

	var total int64
	err := l.svcCtx.DB.WithContext(l.ctx).
		Model(&model.Comment{}).
		Where("post_id = ? AND parent_id = 0 AND deleted_at IS NULL", postId).
		Count(&total).Error
	if err == nil {
		cacheCommentCount(l.ctx, l.svcCtx, postId, total)
	}
	return total, err
}
