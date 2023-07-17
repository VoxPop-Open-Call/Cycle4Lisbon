package models

import (
	"time"

	"github.com/google/uuid"
)

type FCMToken struct {
	BaseModel
	Token        string    `json:"token" gorm:"unique;not null;"`
	Failures     uint16    `json:"-" gorm:"not null;default:0;"`
	LastActiveAt time.Time `json:"-" gorm:"not null;default:now();"`

	UserID uuid.UUID `json:"userId" gorm:"not null;"`
	User   User      `json:"-" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}
