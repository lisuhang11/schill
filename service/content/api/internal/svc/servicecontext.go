// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"SChill/service/content/api/internal/config"
	"SChill/service/content/rpc/contentcenter"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config     config.Config
	ContentRpc contentcenter.ContentCenter
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:     c,
		ContentRpc: contentcenter.NewContentCenter(zrpc.MustNewClient(c.ContentRpc)),
	}
}
