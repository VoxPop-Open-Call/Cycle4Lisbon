package controllers

import (
	"fmt"
	"time"

	"bitbucket.org/pensarmais/cycleforlisbon/src/database/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type FCMTokenController struct {
	db *gorm.DB
}

type RegisterFCMTokenParams struct {
	Token string `json:"token" binding:"required"`
}

// Register or refresh an FCM Token.
//
//	@Summary	Register or refresh a Firebase Cloud Messaging token
//	@Tags		fcm
//	@Produce	json
//	@Security	OIDCToken
//	@Security	AuthHeader
//	@Param		params		body		RegisterFCMTokenParams	true	"Params"
//	@Success	201			{object}	models.FCMToken
//	@Failure	400,401,500	{object}	middleware.ApiError
//	@Router		/fcm/register [post]
func (c *FCMTokenController) Register(
	params RegisterFCMTokenParams,
	ctx *gin.Context,
) (models.FCMToken, error) {
	user, err := tokenUser(ctx, c.db)
	if err != nil {
		return models.FCMToken{},
			fmt.Errorf("failed to retrieve user from token claims: %v", err)
	}

	token := models.FCMToken{
		Token:  params.Token,
		UserID: user.ID,
	}

	err = c.db.
		Clauses(
			clause.OnConflict{
				Columns: []clause.Column{{Name: "token"}},
				DoUpdates: clause.Assignments(map[string]any{
					"last_active_at": time.Now(),
					"updated_at":     time.Now(),
					"failures":       0,
				}),
			},
			clause.Returning{},
		).
		Create(&token).Error

	return token, err
}
