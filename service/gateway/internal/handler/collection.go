package handler

import (
	"net/http"

	"SChill/service/content/rpc/contentcenter"
	"SChill/service/gateway/internal/svc"
	"SChill/service/interaction/rpc/interactioncenter"
)

// GetMyCollectionsHandler returns the current user's collected posts with pagination.
func GetMyCollectionsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, valid := currentUserID(r)
		if !valid {
			fail(w, http.StatusUnauthorized, "请先登录")
			return
		}

		page := pageParam(r, "page", 1)
		pageSize := pageParam(r, "pageSize", 20)

		// 1. Get collected post IDs from interaction service
		collectionsResp, err := svcCtx.InteractionRpc.ListUserCollections(r.Context(), &interactioncenter.ListUserCollectionsReq{
			UserId:   userID,
			Page:     page,
			PageSize: pageSize,
		})
		if err != nil {
			rpcFail(w, err)
			return
		}

		if len(collectionsResp.PostIds) == 0 {
			ok(w, map[string]interface{}{
				"total": collectionsResp.Total,
				"list":  []interface{}{},
			})
			return
		}

		// 2. Batch get post details from content service
		postsResp, err := svcCtx.ContentRpc.BatchGetPost(r.Context(), &contentcenter.BatchGetPostReq{
			PostIds: collectionsResp.PostIds,
		})
		if err != nil {
			rpcFail(w, err)
			return
		}

		ok(w, map[string]interface{}{
			"total": collectionsResp.Total,
			"list":  postsResp.Posts,
		})
	}
}
