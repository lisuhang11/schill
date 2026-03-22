package svc

import (
	"SChill/service/user/rpc/internal/config"
	"SChill/service/user/rpc/internal/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config config.Config
	DB     *gorm.DB
}

func NewServiceContext(c config.Config) *ServiceContext {
	db, err := gorm.Open(mysql.Open(c.Mysql.DataSource), &gorm.Config{})
	if err != nil {
		panic("数据库连接失败: " + err.Error())
	}

	// 自动迁移（仅开发/测试阶段，生产环境建议手动管理）
	err = db.AutoMigrate(&model.User{}, &model.UserProfile{}, &model.UserStat{})
	if err != nil {
		panic("数据库表迁移失败: " + err.Error())
	}

	return &ServiceContext{
		Config: c,
		DB:     db,
	}
}
