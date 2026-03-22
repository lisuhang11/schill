package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ UserStatModel = (*customUserStatModel)(nil)

type (
	// UserStatModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUserStatModel.
	UserStatModel interface {
		userStatModel
		withSession(session sqlx.Session) UserStatModel
	}

	customUserStatModel struct {
		*defaultUserStatModel
	}
)

// NewUserStatModel returns a model for the database table.
func NewUserStatModel(conn sqlx.SqlConn) UserStatModel {
	return &customUserStatModel{
		defaultUserStatModel: newUserStatModel(conn),
	}
}

func (m *customUserStatModel) withSession(session sqlx.Session) UserStatModel {
	return NewUserStatModel(sqlx.NewSqlConnFromSession(session))
}
