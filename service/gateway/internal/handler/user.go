package handler

import (
	"net/http"

	"SChill/service/gateway/internal/svc"
	"SChill/service/gateway/internal/types"
	"SChill/service/user/rpc/usercenter"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetUserInfoHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, valid := parseUintParam(r, "id")
		if !valid {
			fail(w, http.StatusBadRequest, "invalid user id")
			return
		}
		resp, err := svcCtx.UserRpc.GetUserInfo(r.Context(), &usercenter.GetUserInfoReq{UserId: userID})
		if err != nil {
			rpcFail(w, err)
			return
		}
		ok(w, resp)
	}
}

func UpdateProfileHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, valid := currentUserID(r)
		if !valid {
			fail(w, http.StatusUnauthorized, "missing X-User-Id")
			return
		}

		var req types.UpdateProfileReq
		if err := httpx.Parse(r, &req); err != nil {
			fail(w, http.StatusBadRequest, err.Error())
			return
		}

		resp, err := svcCtx.UserRpc.UpdateUserProfileInfo(r.Context(), &usercenter.UpdateUserProfileInfoReq{
			UserId: userID,
			UserProfile: &usercenter.UserProfile{
				UserId:    userID,
				Gender:    req.Gender,
				Birthday:  req.Birthday,
				Signature: req.Signature,
				Location:  req.Location,
				Website:   req.Website,
				Company:   req.Company,
				JobTitle:  req.JobTitle,
				Education: req.Education,
			},
		})
		if err != nil {
			rpcFail(w, err)
			return
		}
		ok(w, resp)
	}
}

func UpdateAvatarHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, valid := currentUserID(r)
		if !valid {
			fail(w, http.StatusUnauthorized, "missing X-User-Id")
			return
		}

		var req types.UpdateAvatarReq
		if err := httpx.Parse(r, &req); err != nil {
			fail(w, http.StatusBadRequest, err.Error())
			return
		}
		resp, err := svcCtx.UserRpc.UpdateAvatar(r.Context(), &usercenter.UpdateAvatarReq{
			UserId:    userID,
			AvatarUrl: req.AvatarUrl,
		})
		if err != nil {
			rpcFail(w, err)
			return
		}
		ok(w, resp)
	}
}
