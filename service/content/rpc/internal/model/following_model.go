package model

type Following struct {
	UserID   uint64 `gorm:"column:user_id" json:"userId"`
	FollowID uint64 `gorm:"column:follow_id" json:"followId"`
}

func (Following) TableName() string {
	return "following"
}
