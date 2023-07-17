package query

import (
	"time"

	"bitbucket.org/pensarmais/cycleforlisbon/src/database/models"
	"gorm.io/gorm"
)

type passwordResetCodes struct{}

var PasswordResetCodes passwordResetCodes

// DeleteOlderThan deletes all password reset codes from the database that have
// been expired for longer than t.
func (passwordResetCodes) DeleteOlderThan(
	t time.Duration, db *gorm.DB,
) (int, error) {
	minExpiryDate := time.Now().Add(-t)
	result := db.Delete(&models.PasswordResetCode{},
		"expires_at < ?", minExpiryDate)

	return int(result.RowsAffected), result.Error
}
