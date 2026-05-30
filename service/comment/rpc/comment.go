package main

import (
	"flag"
	"fmt"

	"SChill/service/comment/rpc/internal/config"
	"SChill/service/comment/rpc/internal/consumers"
	"SChill/service/comment/rpc/internal/server"
	"SChill/service/comment/rpc/internal/svc"
	"SChill/service/comment/rpc/pb"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/comment-rpc.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	svcCtx := svc.NewServiceContext(c)

	serverInstance := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		pb.RegisterCommentCenterServer(grpcServer, server.NewCommentCenterServer(svcCtx))
		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer serverInstance.Stop()

	group := service.NewServiceGroup()
	defer group.Stop()
	group.Add(serverInstance)
	group.Add(consumers.NewCommentConsumerService(svcCtx.CommentConsumer))

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	group.Start()
}
