// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"SChill/service/user/api/internal/config"
	"SChill/service/user/rpc/usercenter"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config  config.Config
	UserRpc usercenter.UserCenter
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:  c,
		UserRpc: usercenter.NewUserCenter(zrpc.MustNewClient(c.UserRpc)),
	}
}
