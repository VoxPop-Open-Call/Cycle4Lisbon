package query

import (
	"bitbucket.org/pensarmais/cycleforlisbon/src/database/models"
	"bitbucket.org/pensarmais/cycleforlisbon/src/server/middleware"
	"gorm.io/gorm"
)

type users struct{}

var Users users

// FromClaims retrieves the user identified by the given claims.
func (users) FromClaims(
	claims *middleware.Claims,
	db *gorm.DB,
) (models.User, error) {
	var user models.User
	err := db.
		Where("subject = ?", claims.Sub).
		Or("subject = ?", claims.Name).
		First(&user).Error

	return user, err
}

func (users) ByEmail(email string, db *gorm.DB) (models.User, error) {
	var user models.User
	err := db.First(&user, "email = ?", email).Error
	return user, err
}

func (users) UpdateStats(
	user *models.User,
	dist, credits float64,
	tx *gorm.DB,
) error {
	user.TripCount += 1
	user.TotalDist += dist
	user.Credits += credits

	return tx.Save(&user).Error
}

// InitiativeCount returns the number of unique initiatives helped by the user.
func (users) InitiativeCount(userID string, db *gorm.DB) (int64, error) {
	var res int64
	err := db.Model(&models.Trip{}).
		Select("initiative_id").
		Where("is_valid = true").
		Where("user_id = ?", userID).
		Group("initiative_id").
		Count(&res).Error

	return res, err
}
