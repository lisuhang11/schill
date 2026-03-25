// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package handler

import (
	"net/http"

	"SChill/service/relation/api/internal/logic"
	"SChill/service/relation/api/internal/svc"
	"SChill/service/relation/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func UnfollowHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UnfollowReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewUnfollowLogic(r.Context(), svcCtx)
		resp, err := l.Unfollow(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
