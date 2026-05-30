package logic

import (
	"crypto/md5"
	"encoding/hex"
	"strconv"
)

func getCacheKey(keyword string, page, pageSize int64) string {
	data := keyword + ":" + strconv.FormatInt(page, 10) + ":" + strconv.FormatInt(pageSize, 10)
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}
