// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package handler

import (
	"net/http"

	"SChill/service/comment/api/internal/logic"
	"SChill/service/comment/api/internal/svc"
	"SChill/service/comment/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func CreateCommentHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CreateCommentReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewCreateCommentLogic(r.Context(), svcCtx)
		resp, err := l.CreateComment(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
