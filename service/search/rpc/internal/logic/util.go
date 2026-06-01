package logic

import (
	"crypto/md5"
	"encoding/hex"
	"strconv"
)

func normalizePage(page, pageSize int64) (int64, int64) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

func getCacheKey(keyword string, page, pageSize int64) string {
	data := keyword + ":" + strconv.FormatInt(page, 10) + ":" + strconv.FormatInt(pageSize, 10)
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}
