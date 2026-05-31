package handler

import (
	"net/http"

	"SChill/service/feed/rpc/feedcenter"
	"SChill/service/gateway/internal/svc"
)

func GetFeedHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		currentUserID, _ := currentUserID(r)
		resp, err := svcCtx.FeedRpc.GetFeedList(r.Context(), &feedcenter.GetFeedListReq{
			FeedType:      r.URL.Query().Get("feedType"),
			Page:          pageParam(r, "page", 1),
			PageSize:      pageParam(r, "pageSize", 20),
			CurrentUserId: currentUserID,
		})
		if err != nil {
			rpcFail(w, err)
			return
		}
		ok(w, resp)
	}
}
