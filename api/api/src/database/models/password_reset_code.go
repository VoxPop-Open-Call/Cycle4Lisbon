package models

import "time"

type PasswordResetCode struct {
	Code      string    `gorm:"primaryKey;type:char(32)"`
	Email     string    `gorm:"not null"`
	CreatedAt time.Time `gorm:"not null"`
	ExpiresAt time.Time `gorm:"not null"`
	Used      bool      `gorm:"not null;default:false"`
}

func (c *PasswordResetCode) IsExpired() bool {
	return time.Now().After(c.ExpiresAt)
}
