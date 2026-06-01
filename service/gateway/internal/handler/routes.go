package handler

import (
	"net/http"

	"SChill/service/gateway/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

func RegisterHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	server.AddRoutes([]rest.Route{
		{Method: http.MethodGet, Path: "/health", Handler: HealthHandler(serverCtx)},
		{Method: http.MethodPost, Path: "/api/auth/register", Handler: RegisterHandler(serverCtx)},
		{Method: http.MethodPost, Path: "/api/auth/login", Handler: LoginHandler(serverCtx)},
		{Method: http.MethodPost, Path: "/api/auth/refresh", Handler: RefreshTokenHandler(serverCtx)},
		{Method: http.MethodGet, Path: "/api/users/:id", Handler: GetUserInfoHandler(serverCtx)},
		{Method: http.MethodPut, Path: "/api/users/me/profile", Handler: UpdateProfileHandler(serverCtx)},
		{Method: http.MethodPut, Path: "/api/users/me/avatar", Handler: UpdateAvatarHandler(serverCtx)},
		{Method: http.MethodGet, Path: "/api/feed", Handler: GetFeedHandler(serverCtx)},
		{Method: http.MethodGet, Path: "/api/posts", Handler: GetPostListHandler(serverCtx)},
		{Method: http.MethodPost, Path: "/api/posts", Handler: CreatePostHandler(serverCtx)},
		{Method: http.MethodGet, Path: "/api/posts/:id", Handler: GetPostDetailHandler(serverCtx)},
		{Method: http.MethodPut, Path: "/api/posts/:id", Handler: UpdatePostHandler(serverCtx)},
		{Method: http.MethodDelete, Path: "/api/posts/:id", Handler: DeletePostHandler(serverCtx)},
		{Method: http.MethodGet, Path: "/api/topics", Handler: GetTopicListHandler(serverCtx)},
		{Method: http.MethodGet, Path: "/api/posts/:id/comments", Handler: GetCommentListHandler(serverCtx)},
		{Method: http.MethodPost, Path: "/api/comments", Handler: CreateCommentHandler(serverCtx)},
		{Method: http.MethodDelete, Path: "/api/comments/:id", Handler: DeleteCommentHandler(serverCtx)},
		{Method: http.MethodPost, Path: "/api/comments/:id/vote", Handler: VoteCommentHandler(serverCtx)},
		{Method: http.MethodPost, Path: "/api/users/:id/follow", Handler: FollowHandler(serverCtx)},
		{Method: http.MethodDelete, Path: "/api/users/:id/follow", Handler: UnfollowHandler(serverCtx)},
		{Method: http.MethodPost, Path: "/api/posts/:id/star", Handler: StarPostHandler(serverCtx)},
		{Method: http.MethodDelete, Path: "/api/posts/:id/star", Handler: UnstarPostHandler(serverCtx)},
		{Method: http.MethodPost, Path: "/api/posts/:id/collect", Handler: CollectPostHandler(serverCtx)},
		{Method: http.MethodDelete, Path: "/api/posts/:id/collect", Handler: UncollectPostHandler(serverCtx)},
		{Method: http.MethodPost, Path: "/api/posts/:id/share", Handler: SharePostHandler(serverCtx)},
		{Method: http.MethodGet, Path: "/api/users/me/collections", Handler: GetMyCollectionsHandler(serverCtx)},
		{Method: http.MethodGet, Path: "/api/search/post", Handler: SearchPostHandler(serverCtx)},
		{Method: http.MethodGet, Path: "/api/search/user", Handler: SearchUserHandler(serverCtx)},
		{Method: http.MethodGet, Path: "/api/search/topic", Handler: SearchTopicHandler(serverCtx)},
	})
}
