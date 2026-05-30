package svc

import (
	"net/http/httputil"
	"net/url"
	"time"

	"SChill/service/comment/rpc/commentcenter"
	"SChill/service/content/rpc/contentcenter"
	"SChill/service/feed/rpc/feedcenter"
	"SChill/service/gateway/internal/config"
	"SChill/service/interaction/rpc/interactioncenter"
	"SChill/service/relation/rpc/relationcenter"
	"SChill/service/user/rpc/usercenter"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config         config.Config
	UserRpc        usercenter.UserCenter
	ContentRpc     contentcenter.ContentCenter
	FeedRpc        feedcenter.FeedCenter
	CommentRpc     commentcenter.CommentCenter
	RelationRpc    relationcenter.RelationCenter
	InteractionRpc interactioncenter.InteractionCenter
	SearchProxy    *httputil.ReverseProxy
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:         c,
		UserRpc:        usercenter.NewUserCenter(zrpc.MustNewClient(c.UserRpc)),
		ContentRpc:     contentcenter.NewContentCenter(zrpc.MustNewClient(c.ContentRpc)),
		FeedRpc:        feedcenter.NewFeedCenter(zrpc.MustNewClient(c.FeedRpc)),
		CommentRpc:     commentcenter.NewCommentCenter(zrpc.MustNewClient(c.CommentRpc)),
		RelationRpc:    relationcenter.NewRelationCenter(zrpc.MustNewClient(c.RelationRpc)),
		InteractionRpc: interactioncenter.NewInteractionCenter(zrpc.MustNewClient(c.InteractionRpc)),
		SearchProxy:    newSearchProxy(c),
	}
}

func newSearchProxy(c config.Config) *httputil.ReverseProxy {
	target, err := url.Parse(c.SearchProxy.Target)
	if err != nil {
		panic(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Transport = &httpTransport{
		dialTimeout:           c.SearchProxy.DialTimeout,
		responseHeaderTimeout: c.SearchProxy.ResponseHeaderTimeout,
	}
	return proxy
}

type httpTransport struct {
	dialTimeout           time.Duration
	responseHeaderTimeout time.Duration
}
