package logic

import (
	"context"
	"strings"

	"SChill/service/canal/internal/model"
	"SChill/service/canal/internal/svc"
)

type PostDocumentBuilder struct {
	svcCtx *svc.ServiceContext
}

type contentRow struct {
	PostID  uint64
	Content string
	Type    int32
}

func NewPostDocumentBuilder(svcCtx *svc.ServiceContext) *PostDocumentBuilder {
	return &PostDocumentBuilder{svcCtx: svcCtx}
}

func (b *PostDocumentBuilder) Build(ctx context.Context, postIDs []uint64) ([]model.ESPost, error) {
	postIDs = uniqueUint64s(postIDs)
	if len(postIDs) == 0 {
		return nil, nil
	}

	type postRow struct {
		ID              uint64
		UserID          uint64
		Title           string
		Cover           string
		CommentCount    int64
		CollectionCount int64
		UpvoteCount     int64
		ShareCount      int64
		Visibility      int32
		IsTop           bool
		IsEssence       bool
		IsLock          bool
		Tags            string
		CreatedAt       int64
		UpdatedAt       int64
		Username        string
		Avatar          string
	}

	var posts []postRow
	if err := b.svcCtx.DB.WithContext(ctx).
		Table("post AS p").
		Select(`p.id, p.user_id, p.title, p.cover, p.comment_count, p.collection_count, p.upvote_count, p.share_count,
			p.visibility, p.is_top, p.is_essence, p.is_lock, p.tags,
			CAST(FLOOR(UNIX_TIMESTAMP(p.created_at) * 1000) AS SIGNED) AS created_at,
			CAST(FLOOR(UNIX_TIMESTAMP(p.updated_at) * 1000) AS SIGNED) AS updated_at,
			u.username, u.avatar`).
		Joins("LEFT JOIN user AS u ON u.user_id = p.user_id AND u.deleted_at IS NULL").
		Where("p.id IN ? AND p.deleted_at IS NULL", postIDs).
		Scan(&posts).Error; err != nil {
		return nil, err
	}

	var contents []contentRow
	if err := b.svcCtx.DB.WithContext(ctx).
		Table("post_content").
		Select("post_id, content, type").
		Where("post_id IN ? AND deleted_at IS NULL", postIDs).
		Order("post_id ASC, sort ASC, id ASC").
		Scan(&contents).Error; err != nil {
		return nil, err
	}

	type topicRow struct {
		PostID    uint64
		TopicName string
	}
	var topics []topicRow
	if err := b.svcCtx.DB.WithContext(ctx).
		Table("post_topic AS pt").
		Select("pt.post_id, t.name AS topic_name").
		Joins("JOIN topic AS t ON t.id = pt.topic_id AND t.deleted_at IS NULL").
		Where("pt.post_id IN ?", postIDs).
		Order("pt.post_id ASC, pt.id ASC").
		Scan(&topics).Error; err != nil {
		return nil, err
	}

	contentMap := make(map[uint64][]contentRow, len(postIDs))
	for _, row := range contents {
		contentMap[row.PostID] = append(contentMap[row.PostID], row)
	}

	topicMap := make(map[uint64][]string, len(postIDs))
	for _, row := range topics {
		name := strings.TrimSpace(row.TopicName)
		if name == "" {
			continue
		}
		topicMap[row.PostID] = append(topicMap[row.PostID], name)
	}

	result := make([]model.ESPost, 0, len(posts))
	for _, post := range posts {
		postContents := contentMap[post.ID]
		title := strings.TrimSpace(post.Title)
		if title == "" {
			title = deriveTitle(postContents)
		}

		cover := strings.TrimSpace(post.Cover)
		if cover == "" {
			cover = deriveCover(postContents)
		}

		contentText := buildContent(postContents)
		result = append(result, model.ESPost{
			ID:              post.ID,
			UserID:          post.UserID,
			Title:           title,
			Cover:           cover,
			Summary:         truncate(contentText, 160),
			Content:         contentText,
			UpdatedAt:       post.UpdatedAt,
			CommentCount:    post.CommentCount,
			CollectionCount: post.CollectionCount,
			UpvoteCount:     post.UpvoteCount,
			ShareCount:      post.ShareCount,
			Visibility:      post.Visibility,
			IsTop:           post.IsTop,
			IsEssence:       post.IsEssence,
			IsLock:          post.IsLock,
			Tags:            splitTags(post.Tags),
			CreatedAt:       post.CreatedAt,
			Username:        post.Username,
			Avatar:          post.Avatar,
			TopicNames:      uniqueStrings(topicMap[post.ID]),
		})
	}

	return result, nil
}

func (b *PostDocumentBuilder) FindPostIDsByUserIDs(ctx context.Context, userIDs []uint64) ([]uint64, error) {
	userIDs = uniqueUint64s(userIDs)
	if len(userIDs) == 0 {
		return nil, nil
	}
	var postIDs []uint64
	err := b.svcCtx.DB.WithContext(ctx).
		Table("post").
		Where("user_id IN ? AND deleted_at IS NULL", userIDs).
		Pluck("id", &postIDs).Error
	return uniqueUint64s(postIDs), err
}

func (b *PostDocumentBuilder) FindPostIDsByTopicIDs(ctx context.Context, topicIDs []uint64) ([]uint64, error) {
	topicIDs = uniqueUint64s(topicIDs)
	if len(topicIDs) == 0 {
		return nil, nil
	}
	var postIDs []uint64
	err := b.svcCtx.DB.WithContext(ctx).
		Table("post_topic").
		Where("topic_id IN ?", topicIDs).
		Pluck("post_id", &postIDs).Error
	return uniqueUint64s(postIDs), err
}

func splitTags(tags string) []string {
	if tags == "" {
		return nil
	}
	parts := strings.Split(tags, ",")
	return uniqueStrings(parts)
}

func deriveTitle(contents []contentRow) string {
	for _, item := range contents {
		text := strings.TrimSpace(item.Content)
		if text == "" {
			continue
		}
		return truncate(text, 48)
	}
	return ""
}

func deriveCover(contents []contentRow) string {
	for _, item := range contents {
		if item.Type == 3 {
			return strings.TrimSpace(item.Content)
		}
	}
	return ""
}

func buildContent(contents []contentRow) string {
	parts := make([]string, 0, len(contents))
	for _, item := range contents {
		if item.Type != 1 && item.Type != 2 {
			continue
		}
		text := strings.TrimSpace(item.Content)
		if text == "" {
			continue
		}
		parts = append(parts, text)
	}
	return strings.Join(parts, "\n")
}

func uniqueUint64s(values []uint64) []uint64 {
	seen := make(map[uint64]struct{}, len(values))
	result := make([]uint64, 0, len(values))
	for _, value := range values {
		if value == 0 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}
	return result
}

func truncate(value string, limit int) string {
	runes := []rune(strings.TrimSpace(value))
	if len(runes) <= limit {
		return string(runes)
	}
	return string(runes[:limit])
}
