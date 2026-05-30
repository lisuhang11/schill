package handler

import (
	"net/http"

	"SChill/service/gateway/internal/svc"
	"SChill/service/interaction/rpc/interactioncenter"
)

func StarPostHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		postID, userID, valid := postAndUser(r)
		if !valid {
			fail(w, http.StatusBadRequest, "invalid post id or missing X-User-Id")
			return
		}
		resp, err := svcCtx.InteractionRpc.StarPost(r.Context(), &interactioncenter.StarPostReq{UserId: userID, PostId: postID})
		if err != nil {
			fail(w, http.StatusBadGateway, err.Error())
			return
		}
		ok(w, resp)
	}
}

func UnstarPostHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		postID, userID, valid := postAndUser(r)
		if !valid {
			fail(w, http.StatusBadRequest, "invalid post id or missing X-User-Id")
			return
		}
		resp, err := svcCtx.InteractionRpc.UnstarPost(r.Context(), &interactioncenter.UnstarPostReq{UserId: userID, PostId: postID})
		if err != nil {
			fail(w, http.StatusBadGateway, err.Error())
			return
		}
		ok(w, resp)
	}
}

func CollectPostHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		postID, userID, valid := postAndUser(r)
		if !valid {
			fail(w, http.StatusBadRequest, "invalid post id or missing X-User-Id")
			return
		}
		resp, err := svcCtx.InteractionRpc.CollectPost(r.Context(), &interactioncenter.CollectPostReq{UserId: userID, PostId: postID})
		if err != nil {
			fail(w, http.StatusBadGateway, err.Error())
			return
		}
		ok(w, resp)
	}
}

func UncollectPostHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		postID, userID, valid := postAndUser(r)
		if !valid {
			fail(w, http.StatusBadRequest, "invalid post id or missing X-User-Id")
			return
		}
		resp, err := svcCtx.InteractionRpc.UncollectPost(r.Context(), &interactioncenter.UncollectPostReq{UserId: userID, PostId: postID})
		if err != nil {
			fail(w, http.StatusBadGateway, err.Error())
			return
		}
		ok(w, resp)
	}
}

func SharePostHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		postID, userID, valid := postAndUser(r)
		if !valid {
			fail(w, http.StatusBadRequest, "invalid post id or missing X-User-Id")
			return
		}
		resp, err := svcCtx.InteractionRpc.SharePost(r.Context(), &interactioncenter.SharePostReq{UserId: userID, PostId: postID})
		if err != nil {
			fail(w, http.StatusBadGateway, err.Error())
			return
		}
		ok(w, resp)
	}
}

func postAndUser(r *http.Request) (uint64, uint64, bool) {
	postID, valid := parseUintParam(r, "id")
	if !valid {
		return 0, 0, false
	}
	userID, valid := currentUserID(r)
	return postID, userID, valid
}
