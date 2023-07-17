package controllers

import (
	"bitbucket.org/pensarmais/cycleforlisbon/src/database/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AchievementController struct {
	db            *gorm.DB
	serverBaseURL string
}

type ListAchievementsFilters struct {
	Pagination
}

type AchievementWithImage struct {
	models.Achievement
	Image string `json:"image"`
}

func achievementWithImage(
	achievement models.Achievement,
	serverBaseURL string,
) AchievementWithImage {
	return AchievementWithImage{
		Achievement: achievement,
		Image:       serverBaseURL + achievement.ImageURI,
	}
}

func achievementsWithImage(
	achievements []models.Achievement,
	serverBaseURL string,
) []AchievementWithImage {
	res := make([]AchievementWithImage, len(achievements))
	for i, achievement := range achievements {
		res[i] = achievementWithImage(achievement, serverBaseURL)
	}
	return res
}

// List all achievements.
//
//	@Summary	List all achievements
//	@Tags		achievements
//	@Produce	json
//	@Security	OIDCToken
//	@Security	AuthHeader
//	@Param		filters		query		ListAchievementsFilters	false	"Filters"
//	@Success	200			{array}		AchievementWithImage
//	@Failure	400,401,500	{object}	middleware.ApiError
//	@Router		/achievements  [get]
func (c *AchievementController) List(
	filters ListAchievementsFilters,
	ctx *gin.Context,
) ([]AchievementWithImage, error) {
	var achievements []models.Achievement
	err := c.db.
		Limit(filters.Limit).
		Offset(filters.Offset).
		Order("code").
		Find(&achievements).Error

	return achievementsWithImage(achievements, c.serverBaseURL), err
}
