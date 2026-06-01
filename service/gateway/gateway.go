package main

import (
	"flag"
	"fmt"
	"net/http"

	"SChill/common/authctx"
	"SChill/service/gateway/internal/config"
	"SChill/service/gateway/internal/handler"
	"SChill/service/gateway/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/gateway.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf, rest.WithCustomCors(func(header http.Header) {
		header.Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
	}, nil, "http://localhost:3000"))
	defer server.Stop()

	server.Use(authctx.OptionalJWTMiddleware(c.Jwt.AccessSecret))

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting gateway at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
