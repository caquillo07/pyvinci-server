package model

import (
	"time"

	"github.com/gofrs/uuid"
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        uuid.UUID
	Username  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (User) TableName() string {
	return "user_record"
}

// CreateUser creates a new user in the database.
//
// NOTE: This method DOES not hash passwords, the caller is responsible for
// password hashing.
func CreateUser(db *gorm.DB, user *User, password string) error {
	type userRecord struct {
		*User
		Password string
	}
	u := userRecord{
		User:     user,
		Password: password,
	}

	if err := db.Create(&u).Error; err != nil {
		return err
	}
	user.ID = u.ID
	return nil
}

func FindUserByUsername(db *gorm.DB, username string) (*User, error) {
	var user User
	if err := db.Where("username = ?", username).Take(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func FindUserByID(db *gorm.DB, id uuid.UUID) (*User, error) {
	var user User
	if err := db.Where("id = ?", id).Take(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (u *User) VerifyPassword(db *gorm.DB, password string) (error, bool) {
	type pass struct {
		Password string
	}
	var p pass
	if err := db.Table("user_record").
		Select("password").
		Where("id = ?", u.ID).
		Take(&p).Error; err != nil {
		return err, false
	}
	err := bcrypt.CompareHashAndPassword([]byte(p.Password), []byte(password))
	return err, err == nil
}
