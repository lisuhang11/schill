package svc

import (
	"SChill/service/content/rpc/contentcenter"
	"SChill/service/feed/rpc/internal/config"
	"SChill/service/interaction/rpc/interactioncenter"
	"SChill/service/relation/rpc/relationcenter"
	"SChill/service/user/rpc/usercenter"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config         config.Config
	ContentRpc     contentcenter.ContentCenter
	UserRpc        usercenter.UserCenter
	RelationRpc    relationcenter.RelationCenter
	InteractionRpc interactioncenter.InteractionCenter
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:         c,
		ContentRpc:     contentcenter.NewContentCenter(zrpc.MustNewClient(c.ContentRpc)),
		UserRpc:        usercenter.NewUserCenter(zrpc.MustNewClient(c.UserRpc)),
		RelationRpc:    relationcenter.NewRelationCenter(zrpc.MustNewClient(c.RelationRpc)),
		InteractionRpc: interactioncenter.NewInteractionCenter(zrpc.MustNewClient(c.InteractionRpc)),
	}
}
