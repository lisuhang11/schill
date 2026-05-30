package main

import (
	"flag"
	"fmt"

	"SChill/service/canal/internal/config"
	"SChill/service/canal/internal/logic"
	"SChill/service/canal/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
)

var configFile = flag.String("f", "etc/canal.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	logx.MustSetup(c.Log)

	svcCtx := svc.NewServiceContext(c)

	// 创建 Canal 同步服务
	canalService := logic.NewCanalService(svcCtx)

	// 使用 service.ServiceGroup 统一管理服务生命周期
	serviceGroup := service.NewServiceGroup()
	defer serviceGroup.Stop()

	// 添加 Canal 同步服务
	serviceGroup.Add(canalService)

	fmt.Printf("Starting canal service...\n")
	serviceGroup.Start()
}
