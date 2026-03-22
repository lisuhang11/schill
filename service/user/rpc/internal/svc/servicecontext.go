package svc

import (
	"SChill/service/user/rpc/internal/config"
	"SChill/service/user/rpc/internal/model"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config           config.Config
	Conn             sqlx.SqlConn // 添加 Conn 字段
	UserModel        model.UserModel
	UserProfileModel model.UserProfileModel
	UserStatModel    model.UserStatModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.Mysql.DataSource)
	return &ServiceContext{
		Config:           c,
		Conn:             conn, // 注入连接
		UserModel:        model.NewUserModel(conn),
		UserProfileModel: model.NewUserProfileModel(conn),
		UserStatModel:    model.NewUserStatModel(conn),
	}
}
