package models

import "gorm.io/gorm"

type Settings struct {
	BaseModel
	// KilometersCreditsRatio is how many kilometers correspond to a credit:
	// `dist / cred = ratio`
	KilometersCreditsRatio float32 `gorm:"not null;type:real;default:1"`
	// CreditsCentsRatio is how many credits correspond to 0.01€:
	// `credits / cents = ratio`
	CreditsCentsRatio float32 `gorm:"not null;type:real;default:0.01"`
}

// Migrate implements the Migrator interface.
// If the Settings table is empty, insert a new record with the default values.
func (Settings) Migrate(db *gorm.DB) error {
	if db.Limit(1).Find(&Settings{}).RowsAffected > 0 {
		return nil
	}

	return db.Create(&Settings{}).Error
}
