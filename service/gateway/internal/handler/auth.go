package handler

import (
	"net/http"

	"SChill/service/gateway/internal/svc"
	"SChill/service/gateway/internal/types"
	"SChill/service/user/rpc/usercenter"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func RegisterHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.RegisterReq
		if err := httpx.Parse(r, &req); err != nil {
			fail(w, http.StatusBadRequest, err.Error())
			return
		}
		resp, err := svcCtx.UserRpc.Register(r.Context(), &usercenter.RegisterReq{
			Username: req.Username,
			Password: req.Password,
		})
		if err != nil {
			fail(w, http.StatusBadGateway, err.Error())
			return
		}
		ok(w, resp)
	}
}

func LoginHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.LoginReq
		if err := httpx.Parse(r, &req); err != nil {
			fail(w, http.StatusBadRequest, err.Error())
			return
		}
		resp, err := svcCtx.UserRpc.Login(r.Context(), &usercenter.LoginReq{
			Username: req.Username,
			Password: req.Password,
		})
		if err != nil {
			fail(w, http.StatusBadGateway, err.Error())
			return
		}
		ok(w, resp)
	}
}

func RefreshTokenHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.RefreshTokenReq
		if err := httpx.Parse(r, &req); err != nil {
			fail(w, http.StatusBadRequest, err.Error())
			return
		}
		resp, err := svcCtx.UserRpc.RefreshToken(r.Context(), &usercenter.RefreshTokenReq{
			RefreshToken: req.RefreshToken,
		})
		if err != nil {
			fail(w, http.StatusBadGateway, err.Error())
			return
		}
		ok(w, resp)
	}
}
