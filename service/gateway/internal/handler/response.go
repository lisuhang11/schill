package handler

import (
	"net/http"
	"strconv"
	"strings"

	"SChill/service/gateway/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
	"github.com/zeromicro/go-zero/rest/pathvar"
)

const currentUserHeader = "X-User-Id"

func ok(w http.ResponseWriter, data interface{}) {
	httpx.OkJson(w, types.Response{Code: 0, Msg: "ok", Data: data})
}

func fail(w http.ResponseWriter, status int, msg string) {
	httpx.WriteJson(w, status, types.Response{Code: int64(status), Msg: msg})
}

func parseUintParam(r *http.Request, name string) (uint64, bool) {
	value := strings.TrimSpace(pathvar.Vars(r)[name])
	id, err := strconv.ParseUint(value, 10, 64)
	return id, err == nil && id > 0
}

func currentUserID(r *http.Request) (uint64, bool) {
	value := strings.TrimSpace(r.Header.Get(currentUserHeader))
	if value == "" {
		return 0, false
	}
	id, err := strconv.ParseUint(value, 10, 64)
	return id, err == nil && id > 0
}

func pageParam(r *http.Request, name string, fallback int64) int64 {
	value := strings.TrimSpace(r.URL.Query().Get(name))
	if value == "" {
		return fallback
	}
	num, err := strconv.ParseInt(value, 10, 64)
	if err != nil || num <= 0 {
		return fallback
	}
	return num
}

func uintQueryParam(r *http.Request, name string, fallback uint64) uint64 {
	value := strings.TrimSpace(r.URL.Query().Get(name))
	if value == "" {
		return fallback
	}
	num, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return fallback
	}
	return num
}
