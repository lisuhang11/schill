package handler

import (
	"net/http"
	"strconv"
	"strings"

	"SChill/common/authctx"
	errutil "SChill/common/error"
	"SChill/service/gateway/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
	"github.com/zeromicro/go-zero/rest/pathvar"
)

func ok(w http.ResponseWriter, data interface{}) {
	httpx.OkJson(w, types.Response{Code: 0, Msg: "ok", Data: data})
}

func fail(w http.ResponseWriter, status int, msg string) {
	httpx.WriteJson(w, status, types.Response{Code: int64(status), Msg: msg})
}

// rpcFail handles gRPC errors by parsing the status code and mapping to HTTP status codes.
func rpcFail(w http.ResponseWriter, err error) {
	code, msg := errutil.ParseRpcError(err)
	httpStatus := gRPCCodeToHTTP(code)
	fail(w, httpStatus, msg)
}

// gRPCCodeToHTTP maps business error codes to HTTP status codes.
func gRPCCodeToHTTP(code int) int {
	switch code {
	case errutil.ErrUnauthorized, errutil.ErrInvalidRefreshToken:
		return http.StatusUnauthorized
	case errutil.ErrInvalidCredentials:
		return http.StatusUnauthorized
	case errutil.ErrInvalidParams, errutil.ErrUsernameOrPasswordEmpty, errutil.ErrPasswordTooWeak,
		errutil.ErrPostTitleEmpty, errutil.ErrPostContentEmpty, errutil.ErrCommentContentEmpty:
		return http.StatusBadRequest
	case errutil.ErrUsernameExists, errutil.ErrAlreadyFollowed, errutil.ErrAlreadyLiked:
		return http.StatusConflict
	case errutil.ErrNoPermission:
		return http.StatusForbidden
	case errutil.ErrUserNotExist, errutil.ErrPostNotExist, errutil.ErrCommentNotExist:
		return http.StatusNotFound
	case errutil.ErrNotFollowing, errutil.ErrNotLiked:
		return http.StatusBadRequest
	case errutil.ErrAccountAbnormal:
		return http.StatusForbidden
	case errutil.ErrTooManyRequests:
		return http.StatusTooManyRequests
	case errutil.ErrCannotFollowSelf:
		return http.StatusBadRequest
	default:
		return http.StatusBadGateway
	}
}

func parseUintParam(r *http.Request, name string) (uint64, bool) {
	value := strings.TrimSpace(pathvar.Vars(r)[name])
	id, err := strconv.ParseUint(value, 10, 64)
	return id, err == nil && id > 0
}

// currentUserID reads the authenticated user ID from the request context,
// populated by the JWT middleware. Returns 0 if no valid JWT token was provided.
func currentUserID(r *http.Request) (uint64, bool) {
	userID := authctx.OptionalUserID(r.Context())
	if userID == 0 {
		return 0, false
	}
	return userID, true
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


