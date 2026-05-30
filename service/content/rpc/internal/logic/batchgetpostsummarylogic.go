package logic

import (
	"context"
	"fmt"
	"strings"
	"time"

	errutil "SChill/common/error"
	"SChill/service/content/rpc/internal/model"
	"SChill/service/content/rpc/internal/svc"
	"SChill/service/content/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

const maxBatchPostSummarySize = 200

type BatchGetPostSummaryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBatchGetPostSummaryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchGetPostSummaryLogic {
	return &BatchGetPostSummaryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *BatchGetPostSummaryLogic) BatchGetPostSummary(in *pb.BatchGetPostSummaryReq) (*pb.BatchGetPostSummaryResp, error) {
	postIDs := uniqueUint64s(in.GetPostIds())
	if len(postIDs) == 0 {
		return &pb.BatchGetPostSummaryResp{Posts: []*pb.PostSummary{}}, nil
	}
	if len(postIDs) > maxBatchPostSummarySize {
		return nil, errutil.RpcBusinessError(errutil.ErrInvalidParams)
	}

	cacheKey := buildBatchPostSummaryLocalCacheKey(postIDs)
	if cached, ok := loadLocalCache[*pb.BatchGetPostSummaryResp](l.svcCtx.LocalCache, cacheKey); ok && cached != nil {
		return cached, nil
	}

	var posts []model.Post
	if err := l.svcCtx.DBRead.WithContext(l.ctx).
		Where("id IN ? AND deleted_at IS NULL", postIDs).
		Find(&posts).Error; err != nil {
		logx.Errorf("batch get post summaries failed: postIds=%v err=%v", postIDs, err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	contentMap, topicMap := l.loadSummaryDependencies(postIDs)
	postMap := make(map[uint64]*pb.PostSummary, len(posts))
	for _, post := range posts {
		contents := contentMap[post.ID]
		title := strings.TrimSpace(post.Title)
		if title == "" {
			title = deriveTitleFromPostContents(contents)
		}

		cover := strings.TrimSpace(post.Cover)
		if cover == "" {
			cover = deriveCoverFromPostContents(contents)
		}

		postMap[post.ID] = &pb.PostSummary{
			Id:         post.ID,
			UserId:     post.UserID,
			Title:      title,
			Cover:      cover,
			Summary:    buildPostSummary(contents),
			TopicNames: topicMap[post.ID],
			Visibility: post.Visibility,
			CreatedAt:  post.CreatedAt.Unix(),
			UpdatedAt:  post.UpdatedAt.Unix(),
		}
	}

	resp := &pb.BatchGetPostSummaryResp{
		Posts: make([]*pb.PostSummary, 0, len(postIDs)),
	}
	for _, postID := range postIDs {
		if item, ok := postMap[postID]; ok {
			resp.Posts = append(resp.Posts, item)
		}
	}
	storeLocalCache(l.svcCtx, cacheKey, resp, time.Minute)

	return resp, nil
}

func (l *BatchGetPostSummaryLogic) loadSummaryDependencies(postIDs []uint64) (map[uint64][]model.PostContent, map[uint64][]string) {
	contentMap := make(map[uint64][]model.PostContent, len(postIDs))
	topicMap := make(map[uint64][]string, len(postIDs))

	var contents []model.PostContent
	if err := l.svcCtx.DBRead.WithContext(l.ctx).
		Where("post_id IN ? AND deleted_at IS NULL", postIDs).
		Order("post_id ASC, sort ASC, id ASC").
		Find(&contents).Error; err != nil {
		logx.Errorf("load post summary contents failed: postIds=%v err=%v", postIDs, err)
	} else {
		for _, content := range contents {
			contentMap[content.PostID] = append(contentMap[content.PostID], content)
		}
	}

	type topicRow struct {
		PostID    uint64 `gorm:"column:post_id"`
		TopicName string `gorm:"column:topic_name"`
	}
	var topicRows []topicRow
	if err := l.svcCtx.DBRead.WithContext(l.ctx).
		Table("post_topic AS pt").
		Select("pt.post_id AS post_id, t.name AS topic_name").
		Joins("JOIN topic AS t ON t.id = pt.topic_id AND t.deleted_at IS NULL").
		Where("pt.post_id IN ?", postIDs).
		Order("pt.post_id ASC, pt.id ASC").
		Scan(&topicRows).Error; err != nil {
		logx.Errorf("load post summary topics failed: postIds=%v err=%v", postIDs, err)
	} else {
		for _, row := range topicRows {
			topicMap[row.PostID] = append(topicMap[row.PostID], row.TopicName)
		}
		for postID, names := range topicMap {
			topicMap[postID] = normalizeTopicNames(names)
		}
	}

	return contentMap, topicMap
}

func buildBatchPostSummaryLocalCacheKey(postIDs []uint64) string {
	if len(postIDs) == 0 {
		return "post:summary:batch:empty"
	}

	parts := make([]string, 0, len(postIDs))
	for _, postID := range postIDs {
		parts = append(parts, uint64ToString(postID))
	}

	return fmt.Sprintf("post:summary:batch:%s", strings.Join(parts, ","))
}

func buildPostSummary(contents []model.PostContent) string {
	if len(contents) == 0 {
		return ""
	}

	parts := make([]string, 0, len(contents))
	for _, content := range contents {
		if content.Type != 1 && content.Type != 2 {
			continue
		}
		text := strings.TrimSpace(content.Content)
		if text == "" {
			continue
		}
		parts = append(parts, text)
		if len(strings.Join(parts, " ")) >= 160 {
			break
		}
	}

	summary := strings.TrimSpace(strings.Join(parts, " "))
	if summary == "" {
		return ""
	}

	runes := []rune(summary)
	if len(runes) > 160 {
		return string(runes[:160])
	}
	return summary
}
