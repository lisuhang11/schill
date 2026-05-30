package model

import "time"

// UserProfile 用户扩展信息表
type UserProfile struct {
	ID        uint64     `gorm:"column:id;primaryKey;autoIncrement"`
	UserID    uint64     `gorm:"column:user_id;uniqueIndex:uk_user_id;not null;comment:用户ID"`
	Gender    int8       `gorm:"column:gender;default:0;comment:性别：0未知，1男，2女"`
	Birthday  *time.Time `gorm:"column:birthday;type:date;comment:生日"` // 可为空
	Signature string     `gorm:"column:signature;size:255;not null;default:'';comment:个性签名"`
	Location  string     `gorm:"column:location;size:64;not null;default:'';comment:所在地"`
	Website   string     `gorm:"column:website;size:255;not null;default:'';comment:个人网站"`
	Company   string     `gorm:"column:company;size:64;not null;default:'';comment:公司"`
	JobTitle  string     `gorm:"column:job_title;size:64;not null;default:'';comment:职位"`
	Education string     `gorm:"column:education;size:64;not null;default:'';comment:教育背景"`
	CreatedAt time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time  `gorm:"column:updated_at;autoUpdateTime"`
}

// TableName 指定表名
func (UserProfile) TableName() string {
	return "user_profile"
}
