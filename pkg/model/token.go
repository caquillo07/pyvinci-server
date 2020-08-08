package model

import (
	"time"

	"github.com/gofrs/uuid"
	"github.com/jinzhu/gorm"
)

type Token struct {
	ID        uuid.UUID
	Valid     bool
	Token     string
	UserID    uuid.UUID `gorm:"column:user_record"`
	CreatedAt time.Time
	UpdatedAt time.Time
	User      *User `gorm:"foreignkey:user_record"`
}

func CreateToken(db *gorm.DB, userID uuid.UUID, token string) error {
	t := Token{
		Valid:  true,
		Token:  token,
		UserID: userID,
	}
	return db.Create(&t).Error
}

func InvalidateAllTokens(db *gorm.DB, userID uuid.UUID) error {
	return db.Table("token").
		Where("user_record = ? AND valid = true", userID).
		Updates(map[string]interface{}{"valid": false, "updated_at": time.Now()}).
		Error
}
