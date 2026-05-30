package handler

import (
	"net/http"

	"SChill/service/search/api/internal/logic"
	"SChill/service/search/api/internal/svc"
	"SChill/service/search/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func SearchTopicHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SearchTopicReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewSearchTopicLogic(r.Context(), svcCtx)
		resp, err := l.SearchTopic(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
