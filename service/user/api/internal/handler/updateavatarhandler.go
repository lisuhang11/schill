package handler

import (
	"io"
	"net/http"

	"SChill/service/user/api/internal/logic"
	"SChill/service/user/api/internal/svc"
	"SChill/service/user/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func UpdateAvatarHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		file, header, err := r.FormFile("avatar")
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		defer file.Close()

		fileBytes, err := io.ReadAll(file)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		req := &types.UpdateAvatarReq{
			Avatar:     fileBytes,
			AvatarName: header.Filename,
		}

		l := logic.NewUpdateAvatarLogic(r.Context(), svcCtx)
		resp, err := l.UpdateAvatar(req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
