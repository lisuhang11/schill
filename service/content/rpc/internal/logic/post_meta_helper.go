package logic

import (
	"strconv"
	"strings"

	"SChill/service/content/rpc/internal/model"
	"SChill/service/content/rpc/pb"
)

type postDerivedMeta struct {
	Title string
	Cover string
}

func normalizeTopicNames(topicNames []string) []string {
	if len(topicNames) == 0 {
		return nil
	}

	result := make([]string, 0, len(topicNames))
	seen := make(map[string]struct{}, len(topicNames))
	for _, name := range topicNames {
		trimmed := strings.TrimSpace(name)
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

func normalizePostContents(items []*pb.PostContentItem) []*pb.PostContentItem {
	result := make([]*pb.PostContentItem, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		content := strings.TrimSpace(item.Content)
		if content == "" {
			continue
		}
		result = append(result, &pb.PostContentItem{
			Type:    item.Type,
			Content: content,
			Sort:    item.Sort,
		})
	}
	return result
}

func deriveTitleFromContentItems(items []*pb.PostContentItem) string {
	for _, item := range items {
		if item == nil {
			continue
		}
		if title := deriveTitleFromContentString(item.Content); title != "" {
			return title
		}
	}
	return ""
}

func deriveTitleFromPostContents(contents []model.PostContent) string {
	for _, content := range contents {
		if title := deriveTitleFromContentString(content.Content); title != "" {
			return title
		}
	}
	return ""
}

func deriveTitleFromContentString(content string) string {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return ""
	}
	runes := []rune(trimmed)
	if len(runes) > 48 {
		return string(runes[:48])
	}
	return trimmed
}

func deriveCoverFromContentItems(items []*pb.PostContentItem) string {
	for _, item := range items {
		if item == nil {
			continue
		}
		if item.Type == 3 {
			return strings.TrimSpace(item.Content)
		}
	}
	return ""
}

func deriveCoverFromPostContents(contents []model.PostContent) string {
	for _, content := range contents {
		if content.Type == 3 {
			return strings.TrimSpace(content.Content)
		}
	}
	return ""
}

func uint64ToString(value uint64) string {
	return strconv.FormatUint(value, 10)
}
