package handler

import (
	"net/http"

	"SChill/service/search/api/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

func RegisterHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodGet,
				Path:    "/search/post",
				Handler: SearchPostHandler(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/search/user",
				Handler: SearchUserHandler(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/search/topic",
				Handler: SearchTopicHandler(serverCtx),
			},
		},
	)
}
