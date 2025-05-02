package auth

import (
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	ID       uint   `gorm:"primaryKey"`
	Username string `gorm:"unique;not null"`
	Password string `gorm:"not null"` // bcrypt hash
}

// BeforeCreate bcrypt-hashes password
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	return nil
}

// CheckPassword compares a plain password to the hash
func (u *User) CheckPassword(pw string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(pw))
}
