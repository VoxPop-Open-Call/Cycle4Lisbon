package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BaseModel struct {
	ID        uuid.UUID `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()" example:"45314277-a7a3-41d4-9626-a5f00db330fa"`
	CreatedAt time.Time `json:"createdAt" gorm:"not null" example:"2023-03-30T17:23:57.146262+02:00"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"not null" example:"2023-03-30T17:34:43.497929+02:00"`
}

type Migrator interface {
	Migrate(*gorm.DB) error
}
