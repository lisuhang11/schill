package handler

import (
	"net/http"

	"SChill/service/search/api/internal/logic"
	"SChill/service/search/api/internal/svc"
	"SChill/service/search/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func SearchPostHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SearchPostReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewSearchPostLogic(r.Context(), svcCtx)
		resp, err := l.SearchPost(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
