package svc

import (
	"SChill/service/user/rpc/internal/config"
	"SChill/service/user/rpc/internal/model"
	"github.com/minio/minio-go/v7"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config config.Config
	DB     *gorm.DB
	MinIO  *minio.Client // 添加 MinIO 客户端
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 初始化数据库
	db, err := gorm.Open(mysql.Open(c.Mysql.DataSource), &gorm.Config{})
	if err != nil {
		panic("数据库连接失败: " + err.Error())
	}

	// 自动迁移（开发阶段）
	err = db.AutoMigrate(&model.User{}, &model.UserProfile{}, &model.UserStat{})
	if err != nil {
		panic("数据库表迁移失败: " + err.Error())
	}

	return &ServiceContext{
		Config: c,
		DB:     db,
	}
}
