package query

import (
	"time"

	"bitbucket.org/pensarmais/cycleforlisbon/src/database/models"
	"gorm.io/gorm"
)

type fcmtokens struct{}

var FCMTokens fcmtokens

// Of retrieves the FCM tokens of the user with the given ID.
func (fcmtokens) Of(userID string, db *gorm.DB) ([]string, error) {
	var tokens []string
	err := db.Model(&models.FCMToken{}).
		Select("token").
		Where("user_id = ?", userID).
		Find(&tokens).Error

	return tokens, err
}

// Increments the failure count of a token, identified by its token string.
func (fcmtokens) IncrementFailureCount(token string, db *gorm.DB) error {
	return db.Model(&models.FCMToken{}).
		Where("token = ?", token).
		Update("failures", gorm.Expr("failures + 1")).Error
}

// Resets the number of failures of a token, identified by its token string.
// It also updates the LastActiveAt time.
func (fcmtokens) ResetFailureCount(token string, db *gorm.DB) error {
	return db.Model(&models.FCMToken{}).
		Where("token = ?", token).
		Updates(map[string]any{
			"failures":       0,
			"last_active_at": time.Now(),
		}).Error
}

// Deletes a token, identified by its token string.
func (fcmtokens) Delete(token string, db *gorm.DB) error {
	return db.Delete(&models.FCMToken{}, "token = ?", token).Error
}

// Deletes all tokens that have been inactive for longer than the provided
// threshold.
func (fcmtokens) DeleteAllInactive(
	inactivityThreshold time.Duration,
	db *gorm.DB,
) (int, error) {
	minLastActivity := time.Now().Add(-inactivityThreshold)
	result := db.Delete(&models.FCMToken{}, "last_active_at < ?", minLastActivity)

	return int(result.RowsAffected), result.Error
}

// Retrieves at most 100 tokens that have accumulated more failures than the
// provided threshold.
func (fcmtokens) GetFailed(
	failureThreshold int,
	db *gorm.DB,
) ([]models.FCMToken, error) {
	tokens := make([]models.FCMToken, 100)
	err := db.
		Where("failures > ?", failureThreshold).
		Limit(100).
		Order("created_at").
		Find(&tokens).Error

	return tokens, err
}
