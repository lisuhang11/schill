package logic

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"SChill/common/cacheprotect"
	errutil "SChill/common/error"
	"SChill/common/redis"
	"SChill/service/comment/rpc/internal/model"
	"SChill/service/comment/rpc/internal/svc"
	"SChill/service/comment/rpc/pb"

	goredis "github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
)

var replyListGroup cacheprotect.Group

const (
	replyListLogicalTTL   = time.Duration(redis.CommentExpire) * time.Second
	replyListPhysicalTTL  = replyListLogicalTTL * 2
	replyListEmptyTTL     = time.Duration(redis.CacheNullExpire) * time.Second
	replyListLockTTL      = 10 * time.Second
	replyListWaitInterval = 50 * time.Millisecond
	replyListWaitAttempts = 20
)

type replyListCacheState struct {
	IDs        []uint64
	HasMore    bool
	NextCursor int64
	Entry      *cacheprotect.Entry
}

type GetReplyListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetReplyListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetReplyListLogic {
	return &GetReplyListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetReplyListLogic) GetReplyList(in *pb.GetReplyListReq) (*pb.GetReplyListResp, error) {
	if in.CommentId == 0 {
		return nil, errutil.RpcBusinessError(errutil.ErrInvalidParams)
	}

	pageSize := in.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	cacheState, err := l.getReplyIDsFromRedis(in.CommentId, in.Cursor, pageSize)
	if err != nil {
		logx.Errorf("get reply ids from redis failed: %v", err)
		return l.getReplyListFromDB(in, pageSize)
	}

	if cacheState.Entry != nil && cacheState.Entry.Empty && cacheState.Entry.IsFresh(time.Now()) {
		return &pb.GetReplyListResp{
			Total:      0,
			List:       []*pb.CommentInfo{},
			HasMore:    false,
			NextCursor: 0,
		}, nil
	}

	if len(cacheState.IDs) == 0 {
		if err := l.ensureReplyCache(in.CommentId); err != nil {
			logx.Errorf("ensure reply cache failed: %v", err)
			return l.getReplyListFromDB(in, pageSize)
		}
		cacheState, err = l.getReplyIDsFromRedis(in.CommentId, in.Cursor, pageSize)
		if err != nil {
			return l.getReplyListFromDB(in, pageSize)
		}
		if cacheState.Entry != nil && cacheState.Entry.Empty && cacheState.Entry.IsFresh(time.Now()) {
			return &pb.GetReplyListResp{
				Total:      0,
				List:       []*pb.CommentInfo{},
				HasMore:    false,
				NextCursor: 0,
			}, nil
		}
	} else if cacheState.Entry == nil || !cacheState.Entry.IsFresh(time.Now()) {
		l.refreshReplyCacheAsync(in.CommentId)
	}

	infoMap, err := l.batchGetCommentInfo(cacheState.IDs)
	if err != nil {
		return l.getReplyListFromDB(in, pageSize)
	}
	contentMap, err := l.batchGetCommentContent(cacheState.IDs)
	if err != nil {
		logx.Errorf("batch get reply content failed: %v", err)
	}

	list := l.assembleCommentList(cacheState.IDs, infoMap, contentMap)
	total := int64(0)
	if !(cacheState.Entry != nil && cacheState.Entry.Empty && len(cacheState.IDs) == 0) {
		total, _ = l.getTotalReplyCount(in.CommentId)
	}

	return &pb.GetReplyListResp{
		Total:      total,
		List:       list,
		HasMore:    cacheState.HasMore,
		NextCursor: cacheState.NextCursor,
	}, nil
}

func (l *GetReplyListLogic) batchGetCommentInfo(commentIDs []uint64) (map[uint64]map[string]string, error) {
	ctx := context.Background()
	result := make(map[uint64]map[string]string, len(commentIDs))

	pipe := l.svcCtx.Redis.Pipeline()
	cmds := make(map[uint64]*goredis.MapStringStringCmd, len(commentIDs))
	for _, id := range commentIDs {
		cmds[id] = pipe.HGetAll(ctx, fmt.Sprintf("%s%d", redis.CommentInfoKey, id))
	}
	if _, err := pipe.Exec(ctx); err != nil && err != goredis.Nil {
		logx.Errorf("pipeline reply info failed: %v", err)
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

func (l *GetReplyListLogic) batchGetCommentContent(commentIDs []uint64) (map[uint64]string, error) {
	ctx := context.Background()
	result := make(map[uint64]string, len(commentIDs))

	pipe := l.svcCtx.Redis.Pipeline()
	cmds := make(map[uint64]*goredis.StringCmd, len(commentIDs))
	for _, id := range commentIDs {
		cmds[id] = pipe.Get(ctx, fmt.Sprintf("%s%d", redis.CommentContentKey, id))
	}
	if _, err := pipe.Exec(ctx); err != nil && err != goredis.Nil {
		logx.Errorf("pipeline reply content failed: %v", err)
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

func (l *GetReplyListLogic) assembleCommentList(ids []uint64, infoMap map[uint64]map[string]string, contentMap map[uint64]string) []*pb.CommentInfo {
	list := make([]*pb.CommentInfo, 0, len(ids))
	for _, id := range ids {
		if info, ok := infoMap[id]; ok {
			if comment := l.mapToCommentInfo(id, info, contentMap[id]); comment != nil {
				list = append(list, comment)
			}
		}
	}
	return list
}

func (l *GetReplyListLogic) mapToCommentInfo(id uint64, info map[string]string, content string) *pb.CommentInfo {
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
		level, _ := strconv.ParseInt(val, 10, 32)
		comment.Level = int32(level)
	}
	if val, ok := info["reply_count"]; ok {
		count, _ := strconv.ParseInt(val, 10, 32)
		comment.ReplyCount = int32(count)
	}
	if val, ok := info["like_count"]; ok {
		count, _ := strconv.ParseInt(val, 10, 32)
		comment.LikeCount = int32(count)
	}
	if val, ok := info["dislike_count"]; ok {
		count, _ := strconv.ParseInt(val, 10, 32)
		comment.DislikeCount = int32(count)
	}
	if val, ok := info["created_at"]; ok {
		ts, _ := strconv.ParseInt(val, 10, 64)
		comment.CreatedAt = ts
	}
	return comment
}

func (l *GetReplyListLogic) getReplyIDsFromRedis(commentId uint64, cursor, pageSize int64) (*replyListCacheState, error) {
	ctx := context.Background()
	entry, err := cacheprotect.LoadEntry(ctx, l.svcCtx.Redis, buildReplyListMetaKey(commentId))
	if err != nil {
		entry = nil
	}

	key := buildReplyListKey(commentId)
	maxScore := "+inf"
	minScore := "-inf"
	if cursor > 0 {
		maxScore = fmt.Sprintf("(%d", cursor)
	}

	idStrings, err := l.svcCtx.Redis.ZRevRangeByScore(ctx, key, minScore, maxScore, 0, pageSize+1)
	if err != nil {
		return nil, err
	}

	hasMore := len(idStrings) > int(pageSize)
	if hasMore {
		idStrings = idStrings[:pageSize]
	}
	ids := make([]uint64, 0, len(idStrings))
	for _, idStr := range idStrings {
		id, _ := strconv.ParseUint(idStr, 10, 64)
		ids = append(ids, id)
	}

	var nextCursor int64
	if len(ids) > 0 {
		nextCursor = int64(ids[len(ids)-1])
	}
	return &replyListCacheState{
		IDs:        ids,
		HasMore:    hasMore,
		NextCursor: nextCursor,
		Entry:      entry,
	}, nil
}

func (l *GetReplyListLogic) ensureReplyCache(commentId uint64) error {
	flightKey := fmt.Sprintf("comment:reply:%d", commentId)
	_, err, _ := replyListGroup.Do(flightKey, func() (interface{}, error) {
		cacheState, loadErr := l.getReplyIDsFromRedis(commentId, 0, 1)
		if loadErr == nil && cacheState != nil && cacheState.Entry != nil && cacheState.Entry.IsFresh(time.Now()) {
			return nil, nil
		}
		return nil, l.rebuildReplyCache(commentId)
	})
	return err
}

func (l *GetReplyListLogic) refreshReplyCacheAsync(commentId uint64) {
	go func() {
		bgLogic := NewGetReplyListLogic(context.Background(), l.svcCtx)
		if err := bgLogic.ensureReplyCache(commentId); err != nil {
			logx.Errorf("async refresh reply cache failed: commentId=%d err=%v", commentId, err)
		}
	}()
}

func (l *GetReplyListLogic) rebuildReplyCache(commentId uint64) error {
	ctx := context.Background()
	lockKey := buildReplyListLockKey(commentId)
	acquired, err := cacheprotect.TryLock(ctx, l.svcCtx.Redis, lockKey, replyListLockTTL)
	if err != nil {
		return err
	}
	if !acquired {
		ok, waitErr := cacheprotect.WaitFor(ctx, replyListWaitAttempts, replyListWaitInterval, func() (bool, error) {
			cacheState, loadErr := l.getReplyIDsFromRedis(commentId, 0, 1)
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
		return fmt.Errorf("reply cache rebuild lock not acquired")
	}
	defer cacheprotect.ReleaseLock(context.Background(), l.svcCtx.Redis, lockKey)

	var comments []*model.Comment
	if err := l.svcCtx.DB.WithContext(ctx).
		Where("parent_id = ? AND deleted_at IS NULL", commentId).
		Order("created_at DESC").
		Find(&comments).Error; err != nil {
		return err
	}

	key := buildReplyListKey(commentId)
	metaKey := buildReplyListMetaKey(commentId)
	_ = l.svcCtx.Redis.Del(ctx, key)

	members := make([]redis.Z, 0, len(comments))
	for _, c := range comments {
		members = append(members, redis.Z{Score: float64(c.CreatedAt.Unix()), Member: c.ID})
	}
	if len(members) > 0 {
		if err := l.svcCtx.Redis.ZAdd(ctx, key, members...); err != nil {
			return err
		}
		_, _ = l.svcCtx.Redis.Expire(ctx, key, replyListPhysicalTTL)
		if err := cacheprotect.StoreMarker(ctx, l.svcCtx.Redis, metaKey, replyListLogicalTTL, replyListPhysicalTTL); err != nil {
			return err
		}
		cacheReplyCount(ctx, l.svcCtx, commentId, int64(len(comments)))
	} else {
		if err := cacheprotect.StoreEmpty(ctx, l.svcCtx.Redis, metaKey, replyListEmptyTTL, replyListEmptyTTL); err != nil {
			return err
		}
		cacheReplyCount(ctx, l.svcCtx, commentId, 0)
	}
	return nil
}

func (l *GetReplyListLogic) getReplyListFromDB(in *pb.GetReplyListReq, pageSize int64) (*pb.GetReplyListResp, error) {
	var comments []*model.Comment
	query := l.svcCtx.DB.WithContext(l.ctx).Where("parent_id = ? AND deleted_at IS NULL", in.CommentId)
	if in.Cursor > 0 {
		query = query.Where("id < ?", in.Cursor)
	}
	if err := query.Order("created_at DESC").Limit(int(pageSize + 1)).Find(&comments).Error; err != nil {
		logx.Errorf("query reply list failed: %v", err)
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
		var contents []*model.CommentContent
		if err := l.svcCtx.DB.WithContext(l.ctx).Where("comment_id IN ?", commentIDs).Find(&contents).Error; err != nil {
			return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
		}
		for _, item := range contents {
			contentMap[item.CommentID] = item
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
	total, err := l.getTotalReplyCount(in.CommentId)
	if err != nil {
		logx.Errorf("query total reply count failed: %v", err)
	}

	return &pb.GetReplyListResp{
		Total:      total,
		List:       list,
		HasMore:    hasMore,
		NextCursor: nextCursor,
	}, nil
}

func (l *GetReplyListLogic) getTotalReplyCount(commentId uint64) (int64, error) {
	cacheKey := fmt.Sprintf("%s%d", redis.CommentReplyCountKey, commentId)
	var cached int64
	if err := l.svcCtx.Cache.GetCtx(l.ctx, cacheKey, &cached); err == nil {
		return cached, nil
	}

	var total int64
	err := l.svcCtx.DB.WithContext(l.ctx).
		Model(&model.Comment{}).
		Where("parent_id = ? AND deleted_at IS NULL", commentId).
		Count(&total).Error
	if err == nil {
		cacheReplyCount(l.ctx, l.svcCtx, commentId, total)
	}
	return total, err
}
