package handler

import (
	"net/http"

	"SChill/service/comment/rpc/commentcenter"
	"SChill/service/gateway/internal/svc"
	"SChill/service/gateway/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetCommentListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		postID, valid := parseUintParam(r, "id")
		if !valid {
			fail(w, http.StatusBadRequest, "invalid post id")
			return
		}
		resp, err := svcCtx.CommentRpc.GetCommentList(r.Context(), &commentcenter.GetCommentListReq{
			PostId:   postID,
			Cursor:   pageParam(r, "cursor", 0),
			PageSize: pageParam(r, "pageSize", 20),
			SortType: r.URL.Query().Get("sortType"),
		})
		if err != nil {
			fail(w, http.StatusBadGateway, err.Error())
			return
		}
		ok(w, resp)
	}
}

func CreateCommentHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, valid := currentUserID(r)
		if !valid {
			fail(w, http.StatusUnauthorized, "missing X-User-Id")
			return
		}
		var req types.CreateCommentReq
		if err := httpx.Parse(r, &req); err != nil {
			fail(w, http.StatusBadRequest, err.Error())
			return
		}
		resp, err := svcCtx.CommentRpc.CreateComment(r.Context(), &commentcenter.CreateCommentReq{
			UserId:        userID,
			PostId:        req.PostId,
			ParentId:      req.ParentId,
			ReplyToUserId: req.ReplyToUserId,
			Content:       req.Content,
		})
		if err != nil {
			fail(w, http.StatusBadGateway, err.Error())
			return
		}
		ok(w, resp)
	}
}

func DeleteCommentHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		commentID, valid := parseUintParam(r, "id")
		if !valid {
			fail(w, http.StatusBadRequest, "invalid comment id")
			return
		}
		userID, valid := currentUserID(r)
		if !valid {
			fail(w, http.StatusUnauthorized, "missing X-User-Id")
			return
		}
		resp, err := svcCtx.CommentRpc.DeleteComment(r.Context(), &commentcenter.DeleteCommentReq{
			CommentId: commentID,
			UserId:    userID,
		})
		if err != nil {
			fail(w, http.StatusBadGateway, err.Error())
			return
		}
		ok(w, resp)
	}
}

func VoteCommentHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		commentID, valid := parseUintParam(r, "id")
		if !valid {
			fail(w, http.StatusBadRequest, "invalid comment id")
			return
		}
		userID, valid := currentUserID(r)
		if !valid {
			fail(w, http.StatusUnauthorized, "missing X-User-Id")
			return
		}
		var req types.VoteCommentReq
		if err := httpx.Parse(r, &req); err != nil {
			fail(w, http.StatusBadRequest, err.Error())
			return
		}
		resp, err := svcCtx.CommentRpc.VoteComment(r.Context(), &commentcenter.VoteCommentReq{
			CommentId: commentID,
			UserId:    userID,
			VoteType:  req.VoteType,
		})
		if err != nil {
			fail(w, http.StatusBadGateway, err.Error())
			return
		}
		ok(w, resp)
	}
}
