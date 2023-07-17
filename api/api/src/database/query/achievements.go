package query

import (
	"time"

	"bitbucket.org/pensarmais/cycleforlisbon/src/database/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type achievements struct{}

var Achievements achievements

func (achievements) All(db *gorm.DB) ([]models.Achievement, error) {
	var achs []models.Achievement
	err := db.Model(&models.Achievement{}).
		Order("code").
		Find(&achs).Error
	return achs, err
}

func (achievements) getOrCreate(
	userID uuid.UUID,
	code string,
	tx *gorm.DB,
) (models.UserAchievement, error) {
	ach := models.UserAchievement{
		AchievementCode: code,
		UserID:          userID,
	}
	err := tx.Model(&models.UserAchievement{}).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		FirstOrCreate(&ach).Error

	return ach, err
}

func (achievements) Set(
	userID uuid.UUID,
	code string,
	state bool,
	tx *gorm.DB,
) (models.UserAchievement, error) {
	ach, err := Achievements.getOrCreate(userID, code, tx)
	if err != nil {
		return models.UserAchievement{}, err
	}

	if !ach.Achieved && state {
		now := time.Now()
		ach.AchievedAt = &now
	}
	ach.Achieved = state

	err = tx.Save(&ach).Error
	return ach, err
}

func (achievements) SetCompletion(
	userID uuid.UUID,
	code string,
	value float64,
	tx *gorm.DB,
) (models.UserAchievement, error) {
	ach, err := Achievements.getOrCreate(userID, code, tx)
	if err != nil {
		return models.UserAchievement{}, err
	}

	ach.Completion = value

	err = tx.Save(&ach).Error
	return ach, err
}

func (achievements) Get(
	userID uuid.UUID,
	code string,
	db *gorm.DB,
) (models.UserAchievement, error) {
	var ach models.UserAchievement
	err := db.Model(&models.UserAchievement{}).
		Where("achievement_code = ?", code).
		Where("user_id = ?", userID).
		Find(&ach).Error
	return ach, err
}

func (achievements) List(
	userID string,
	db *gorm.DB,
) ([]models.UserAchievement, error) {
	var achs []models.UserAchievement
	err := db.Model(&models.UserAchievement{}).
		Where("user_id = ?", userID).
		Joins("Achievement").
		Order("achievement_code").
		Find(&achs).Error
	return achs, err
}
