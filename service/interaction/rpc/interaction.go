package main

import (
	"flag"
	"fmt"

	"SChill/service/interaction/rpc/internal/config"
	"SChill/service/interaction/rpc/internal/logic"
	"SChill/service/interaction/rpc/internal/server"
	"SChill/service/interaction/rpc/internal/svc"
	"SChill/service/interaction/rpc/pb"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/interaction-rpc.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)

	rpcServer := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		pb.RegisterInteractionCenterServer(grpcServer, server.NewInteractionCenterServer(ctx))
		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer rpcServer.Stop()

	group := service.NewServiceGroup()
	defer group.Stop()
	group.Add(rpcServer)
	group.Add(logic.NewInteractionConsumerService(logic.NewInteractionConsumer(ctx)))

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	group.Start()
}
