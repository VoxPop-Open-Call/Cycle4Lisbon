package models

import (
	"bitbucket.org/pensarmais/cycleforlisbon/src/database/types"
	"github.com/google/uuid"
)

type User struct {
	BaseModel
	Profile
	Subject        string `json:"subject" gorm:"unique;not null"`
	Email          string `json:"email" gorm:"unique;not null"`
	HashedPassword string `json:"-" gorm:"type:varchar(60);default:null"`
	Verified       bool   `json:"verified" gorm:"not null;default:false"`
	Admin          bool   `json:"-" gorm:"not null;default:false"`

	// TripCount is the total number of valid trips.
	TripCount uint `json:"tripCount" gorm:"type:integer;not null;default:0"`

	// TotalDist is the sum of the distances of the user's valid trips, in
	// kilometers.
	TotalDist float64 `json:"totalDist" gorm:"not null;default:0"`

	// Credits is the total number of credits earned by the user.
	Credits float64 `json:"credits" gorm:"not null;default:0"`

	InitiativeID *uuid.UUID  `json:"initiativeId,omitempty" gorm:"default:null"`
	Initiative   *Initiative `json:"initiative,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type Profile struct {
	Name     string      `json:"name,omitempty" gorm:"default:null"`
	Username string      `json:"username" binding:"required" gorm:"unique;default:null"`
	Gender   string      `json:"gender,omitempty" gorm:"type:varchar(1);default:null" binding:"omitempty,oneof=M F X"`
	Birthday *types.Date `json:"birthday,omitempty" gorm:"default:null"`
}
