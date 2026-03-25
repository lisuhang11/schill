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

func GetReplyListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetReplyListReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewGetReplyListLogic(r.Context(), svcCtx)
		resp, err := l.GetReplyList(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
