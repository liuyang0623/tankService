package users

import "gorm.io/gorm"

// User represents a WeChat Mini Program user in the database.
type User struct {
	gorm.Model
	Openid     string `gorm:"type:varchar(191);uniqueIndex;not null" json:"openid"`
	Unionid    string `gorm:"type:varchar(100);default:''" json:"unionid,omitempty"`
	SessionKey string `gorm:"type:varchar(100);default:''" json:"-"`
	Nickname   string `gorm:"type:varchar(191);not null;default:''" json:"nickname"`
	Avatar     string `gorm:"type:varchar(500);not null;default:''" json:"avatar"`
	Phone      string `gorm:"type:varchar(20);default:''" json:"phone,omitempty"`
	Bio        string `gorm:"type:varchar(191);default:''" json:"bio"`
	Gender     int    `gorm:"default:0" json:"gender"`
}

// TableName returns the database table name for User.
func (User) TableName() string {
	return "users"
}
