// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package handler

import (
	"net/http"

	"SChill/service/user/api/internal/logic"
	"SChill/service/user/api/internal/svc"
	"SChill/service/user/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func UpdateUserProfileInfoHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UpdateUserProfileInfoReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewUpdateUserProfileInfoLogic(r.Context(), svcCtx)
		resp, err := l.UpdateUserProfileInfo(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
