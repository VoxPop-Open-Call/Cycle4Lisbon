package query

import (
	"bitbucket.org/pensarmais/cycleforlisbon/src/database/models"
	"gorm.io/gorm"
)

type settings struct{}

var Settings settings

func (settings) KilometersCreditsRatio(db *gorm.DB) (float32, error) {
	var result float32
	err := db.Model(&models.Settings{}).
		Select("kilometers_credits_ratio").
		First(&result).Error
	return result, err
}
