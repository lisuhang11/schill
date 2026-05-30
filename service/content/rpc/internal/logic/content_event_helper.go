package logic

import (
	"context"
	"strings"
	"time"

	"SChill/common/mq"
	"SChill/service/content/rpc/internal/model"
	"SChill/service/content/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

const contentTypePost = "post"

type contentSummaryItem struct {
	Type    int32
	Content string
}

func publishContentChangedEvent(ctx context.Context, svcCtx *svc.ServiceContext, post *model.Post, opType, summary string, topicNames []string) {
	_ = ctx
	if svcCtx == nil || post == nil || strings.TrimSpace(svcCtx.Config.KqPusherConf.TopicChanged) == "" {
		return
	}

	payload := mq.ContentChangedEvent{
		ContentID:   post.ID,
		AuthorID:    post.UserID,
		OpType:      strings.TrimSpace(opType),
		ContentType: contentTypePost,
		Visibility:  post.Visibility,
		Status:      contentStatusFromDeleted(post.DeletedAt),
		Title:       strings.TrimSpace(post.Title),
		Summary:     strings.TrimSpace(summary),
		Cover:       strings.TrimSpace(post.Cover),
		TopicNames:  normalizeTopicNames(topicNames),
		Tags:        splitTagList(post.Tags),
		CreatedAt:   post.CreatedAt.UnixMilli(),
		UpdatedAt:   post.UpdatedAt.UnixMilli(),
	}

	eventType := "content." + payload.OpType
	msg, err := mq.BuildEnvelopeProducerMessage(
		svcCtx.Config.KqPusherConf.TopicChanged,
		uint64ToString(post.ID),
		eventType,
		"content-rpc",
		"content",
		uint64ToString(post.ID),
		payload,
	)
	if err != nil {
		logx.Errorf("build content changed event failed: postId=%d err=%v", post.ID, err)
		return
	}

	go func() {
		if _, _, err := svcCtx.KafkaProducer.SendMessage(msg); err != nil {
			logx.Errorf("send content changed event failed: postId=%d op=%s err=%v", post.ID, payload.OpType, err)
		}
	}()
}

func buildContentSummaryFromItems(items []contentSummaryItem) string {
	parts := make([]string, 0, len(items))
	for _, item := range items {
		if item.Type != 1 && item.Type != 2 {
			continue
		}

		text := strings.TrimSpace(item.Content)
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

func splitTagList(tags string) []string {
	if strings.TrimSpace(tags) == "" {
		return nil
	}

	parts := strings.Split(tags, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		result = append(result, trimmed)
	}

	return result
}

func contentStatusFromDeleted(deletedAt *time.Time) string {
	if deletedAt != nil {
		return "deleted"
	}

	return "online"
}
