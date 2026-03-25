package svc

import (
	"SChill/service/content/rpc/internal/config"
	"SChill/service/content/rpc/internal/model"
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

	err = db.AutoMigrate(&model.Post{}, &model.PostContent{}, &model.PostTopic{})
	if err != nil {
		panic("数据库表迁移失败: " + err.Error())
	}

	return &ServiceContext{
		Config: c,
		DB:     db,
	}
}
