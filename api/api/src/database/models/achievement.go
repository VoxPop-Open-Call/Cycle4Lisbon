package models

import (
	"time"

	"github.com/google/uuid"
)

type Achievement struct {
	Code     string `json:"code" gorm:"type:varchar(20);primaryKey"`
	ImageURI string `json:"imageURI" gorm:"not null"`
	Name     string `json:"name"`
	Desc     string `json:"desc"`
}

type UserAchievement struct {
	UserID uuid.UUID `json:"userID" gorm:"primaryKey;not null"`
	User   User      `json:"-" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	AchievementCode string      `json:"achievementCode" gorm:"primaryKey;not null"`
	Achievement     Achievement `json:"achievement" gorm:"foreignKey:AchievementCode;references:Code;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	// Completion of the achievement, from 0 to 1.
	Completion float64    `json:"completion" gorm:"not null;default:0"`
	Achieved   bool       `json:"achieved" gorm:"not null;default:false"`
	AchievedAt *time.Time `json:"achievedAt,omitempty" gorm:"default:null"`
}
