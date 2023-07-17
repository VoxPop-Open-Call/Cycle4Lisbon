package seeders

import (
	"bitbucket.org/pensarmais/cycleforlisbon/src/database/models"
	"bitbucket.org/pensarmais/cycleforlisbon/src/util/random"
	"gorm.io/gorm"
)

type fcmtoken struct{}

var FCMToken fcmtoken

func (fcmtoken) Seed(db *gorm.DB) {
	var users []models.User
	db.Limit(5).Find(&users)

	tokens := make([]*models.FCMToken, len(users))
	for i, user := range users {
		token := &models.FCMToken{
			User:  user,
			Token: random.String(50),
		}
		tokens[i] = token
	}

	db.CreateInBatches(tokens, 50)
}
