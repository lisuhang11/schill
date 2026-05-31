package handler

import (
	"net/http"

	"SChill/service/gateway/internal/svc"
	"SChill/service/relation/rpc/relationcenter"
)

func FollowHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		targetUserID, valid := parseUintParam(r, "id")
		if !valid {
			fail(w, http.StatusBadRequest, "invalid target user id")
			return
		}
		userID, valid := currentUserID(r)
		if !valid {
			fail(w, http.StatusUnauthorized, "missing X-User-Id")
			return
		}
		resp, err := svcCtx.RelationRpc.Follow(r.Context(), &relationcenter.FollowReq{
			UserId:       userID,
			TargetUserId: targetUserID,
		})
		if err != nil {
			rpcFail(w, err)
			return
		}
		ok(w, resp)
	}
}

func UnfollowHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		targetUserID, valid := parseUintParam(r, "id")
		if !valid {
			fail(w, http.StatusBadRequest, "invalid target user id")
			return
		}
		userID, valid := currentUserID(r)
		if !valid {
			fail(w, http.StatusUnauthorized, "missing X-User-Id")
			return
		}
		resp, err := svcCtx.RelationRpc.Unfollow(r.Context(), &relationcenter.UnfollowReq{
			UserId:       userID,
			TargetUserId: targetUserID,
		})
		if err != nil {
			rpcFail(w, err)
			return
		}
		ok(w, resp)
	}
}
