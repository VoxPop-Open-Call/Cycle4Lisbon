package models

import (
	"bitbucket.org/pensarmais/cycleforlisbon/src/database/types"
	"github.com/google/uuid"
)

type Initiative struct {
	BaseModel
	Title       string     `json:"title" gorm:"unique;not null"`
	Description string     `json:"description" gorm:"not null"`
	EndDate     types.Date `json:"endDate" gorm:"not null"`

	// Goal is the target number of credits.
	Goal uint32 `json:"goal" gorm:"not null"`
	// Credits is the current credit score.
	Credits float64 `json:"credits" gorm:"nol null;default:0"`

	Enabled bool `json:"enabled" gorm:"not null;default:false"`

	InstitutionID uuid.UUID   `json:"institutionId"`
	Institution   Institution `json:"institution" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	Sponsors []Institution `json:"sponsors,omitempty" gorm:"many2many:initiative_sponsors;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	SDGs []SDG `json:"sdgs" gorm:"many2many:initiative_sdgs;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}
