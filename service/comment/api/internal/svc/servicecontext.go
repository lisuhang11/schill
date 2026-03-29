// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"SChill/service/comment/api/internal/config"
	"SChill/service/comment/rpc/commentcenter"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config     config.Config
	CommentRpc commentcenter.CommentCenter
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:     c,
		CommentRpc: commentcenter.NewCommentCenter(zrpc.MustNewClient(c.CommentRpc)),
	}
}
