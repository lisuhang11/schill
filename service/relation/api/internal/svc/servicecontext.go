// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"SChill/service/relation/api/internal/config"
	"SChill/service/relation/rpc/relationcenter"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config      config.Config
	RelationRpc relationcenter.RelationCenter
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:      c,
		RelationRpc: relationcenter.NewRelationCenter(zrpc.MustNewClient(c.RelationRpc)),
	}
}
