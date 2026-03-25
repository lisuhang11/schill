// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package handler

import (
	"net/http"

	"SChill/service/content/api/internal/logic"
	"SChill/service/content/api/internal/svc"
	"SChill/service/content/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetPostListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetPostListReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewGetPostListLogic(r.Context(), svcCtx)
		resp, err := l.GetPostList(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
