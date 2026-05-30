package minioUtil

import (
	"fmt"
	"time"
)

// 生成唯一对象名
func GenerateMinIOObjectName(userId uint64, mimeType string) string {
	ext := "jpg"
	switch mimeType {
	case "image/jpeg":
		ext = "jpg"
	case "image/png":
		ext = "png"
	case "image/gif":
		ext = "gif"
	}
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("avatars/%d/%d_%d.%s", userId, timestamp, userId, ext)
}
