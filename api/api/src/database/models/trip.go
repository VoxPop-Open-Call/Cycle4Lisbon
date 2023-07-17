package models

import "github.com/google/uuid"

type Trip struct {
	BaseModel

	GPX     []byte `json:"-" gorm:"type:xml;not null"`
	GPXHash []byte `json:"-" gorm:"unique;not null"`

	StartLat  float64 `json:"startLat,omitempty"`  // StartLat is latitude of the starting point in decimal degrees.
	StartLon  float64 `json:"startLon,omitempty"`  // StartLon is longitude of the starting point in decimal degrees.
	EndLat    float64 `json:"endLat,omitempty"`    // EndLat is latitude of the ending point in decimal degrees.
	EndLon    float64 `json:"endLon,omitempty"`    // EndLon is longitude of the ending point in decimal degrees.
	StartAddr string  `json:"startAddr,omitempty"` // StartAddr is the address of the starting point.
	EndAddr   string  `json:"endAddr,omitempty"`   // EndAddr is the address of the ending point.

	// IsValid indicates whether the trip was considered to have been performed
	// on a bicycle.
	//
	// An invalid trip doesn't contribute to the user stats nor the initiative's
	// credit score.
	IsValid bool `json:"isValid" gorm:"not null;default:false"`
	// NotValidReason is the reason why the trip was not considered valid.
	NotValidReason string `json:"notValidReason,omitempty"`

	// Distance is the total distance of the trip, in kilometers.
	Distance float64 `json:"distance" gorm:"not null;default:0"`

	// Credits is the amount of credits awarded to this trip.
	Credits float64 `json:"credits" gorm:"not null;default:0"`

	// Duration is the total duration of the trip, in seconds.
	Duration float64 `json:"duration" gorm:"not null;default:0"`

	// DurationInMotion is the time in seconds spent in motion (where the
	// average speed is greater than a threshold).
	DurationInMotion float64 `json:"durationInMotion" gorm:"not null;default:0"`

	UserID uuid.UUID `json:"userId" gorm:"not null"`
	User   *User     `json:"user,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	InitiativeID *uuid.UUID  `json:"initiativeId,omitempty"`
	Initiative   *Initiative `json:"initiative,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}
