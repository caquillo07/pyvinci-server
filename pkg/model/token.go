package model

import (
	"time"

	"github.com/jinzhu/gorm"
)

type Token struct {
	ID        int
	Valid     bool
	Token     string
	UserID    int `gorm:"column:user_record"`
	CreatedAt time.Time
	UpdatedAt time.Time
	User      *User `gorm:"foreignkey:user_record"`
}

func CreateToken(db *gorm.DB, userID int, token string) error {
	t := Token{
		Valid: true,
		Token: token,
		UserID:  userID,
	}
	return db.Create(&t).Error
}

func InvalidateAllTokens(db *gorm.DB, userID int) error {
	return db.Table("token").
		Where("user_record = ? AND valid = true", userID).
		Updates(map[string]interface{}{"valid": false, "updated_at": time.Now()}).
		Error
}
