package handler

import (
	"context"
	"net/http"

	"SChill/common/authctx"
	"SChill/service/gateway/internal/svc"
	"SChill/service/gateway/internal/types"
	"SChill/service/user/rpc/usercenter"

	"github.com/zeromicro/go-zero/rest/httpx"
	"google.golang.org/grpc/metadata"
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

		ctx := withAuthMetadata(r.Context(), r)
		resp, err := svcCtx.UserRpc.UpdateUserProfileInfo(ctx, &usercenter.UpdateUserProfileInfoReq{
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
		ctx := withAuthMetadata(r.Context(), r)
		resp, err := svcCtx.UserRpc.UpdateAvatar(ctx, &usercenter.UpdateAvatarReq{
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

// withAuthMetadata attaches the JWT access token from the HTTP request to gRPC outgoing metadata.
// This allows the downstream RPC server to verify the caller's identity independently.
func withAuthMetadata(ctx context.Context, r *http.Request) context.Context {
	token := authctx.TokenFromContext(r.Context())
	if token == "" {
		return ctx
	}
	return metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
}
