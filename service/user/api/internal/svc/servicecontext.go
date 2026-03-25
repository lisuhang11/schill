// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"SChill/service/user/api/internal/config"
	"SChill/service/user/rpc/usercenter"
	"context"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config  config.Config
	UserRpc usercenter.UserCenter
	MinIO   *minio.Client // MinIO 客户端
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 初始化 MinIO 客户端
	minioClient, err := minio.New(c.MinIO.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(c.MinIO.AccessKey, c.MinIO.SecretKey, ""),
		Secure: c.MinIO.UseSSL,
	})
	if err != nil {
		panic("MinIO 客户端初始化失败: " + err.Error())
	}

	// 确保 bucket 存在
	ctx := context.Background()
	exists, err := minioClient.BucketExists(ctx, c.MinIO.Bucket)
	if err != nil {
		panic("检查 bucket 失败: " + err.Error())
	}
	if !exists {
		err := minioClient.MakeBucket(ctx, c.MinIO.Bucket, minio.MakeBucketOptions{})
		if err != nil {
			panic("创建 bucket 失败: " + err.Error())
		}
	}

	return &ServiceContext{
		Config:  c,
		UserRpc: usercenter.NewUserCenter(zrpc.MustNewClient(c.UserRpc)),
		MinIO:   minioClient,
	}
}
