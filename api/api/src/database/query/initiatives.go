package query

import (
	"errors"
	"time"

	"bitbucket.org/pensarmais/cycleforlisbon/src/database/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type initiatives struct{}

var Initiatives initiatives

var (
	ErrInitiativeEnded = errors.New("the initiative has reached its goal or expired")
)

// WithAssociations prepares a query for initiatives with the respective
// associations.
func (initiatives) WithAssociations(db *gorm.DB) (tx *gorm.DB) {
	return db.Model(&models.Initiative{}).
		Preload("Institution").
		Preload("Sponsors").
		Preload("SDGs")
}

// Credit adds the given value to the initiative's credit score.
//
// If the initiative is no longer active (has reached the goal or has expired),
// ErrInitiativeEnded is returned.
func (initiatives) Credit(id string, v float64, tx *gorm.DB) error {
	var initiative models.Initiative
	if err := tx.Model(&models.Initiative{}).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", id).
		Find(&initiative).Error; err != nil {
		return err
	}

	if !initiative.Enabled ||
		initiative.Credits >= float64(initiative.Goal) ||
		time.Now().After(initiative.EndDate.Time()) {
		return ErrInitiativeEnded
	}

	initiative.Credits += v
	return tx.Save(initiative).Error
}
