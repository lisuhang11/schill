package handler

import (
	"net/http"
	"strings"

	"SChill/service/gateway/internal/svc"
	"SChill/service/search/rpc/searchcenter"
)

func SearchPostHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := searchReqFromQuery(r)
		resp, err := svcCtx.SearchRpc.SearchPost(r.Context(), req)
		if err != nil {
			rpcFail(w, err)
			return
		}
		ok(w, resp)
	}
}

func SearchUserHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := searchReqFromQuery(r)
		resp, err := svcCtx.SearchRpc.SearchUser(r.Context(), req)
		if err != nil {
			rpcFail(w, err)
			return
		}
		ok(w, resp)
	}
}

func SearchTopicHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := searchReqFromQuery(r)
		resp, err := svcCtx.SearchRpc.SearchTopic(r.Context(), req)
		if err != nil {
			rpcFail(w, err)
			return
		}
		ok(w, resp)
	}
}

func searchReqFromQuery(r *http.Request) *searchcenter.SearchReq {
	return &searchcenter.SearchReq{
		Keyword:  strings.TrimSpace(r.URL.Query().Get("keyword")),
		Page:     pageParam(r, "page", 1),
		PageSize: pageParam(r, "pageSize", 20),
	}
}
