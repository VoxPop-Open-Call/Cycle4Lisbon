package route

import (
	"bitbucket.org/pensarmais/cycleforlisbon/src/server/controllers"
	"bitbucket.org/pensarmais/cycleforlisbon/src/server/handle"
	"github.com/gin-gonic/gin"
)

func Achievements(
	router *gin.RouterGroup,
	auth gin.HandlerFunc,
	store *controllers.Store,
) {
	achievements := router.Group("/achievements", auth)
	{
		achievements.GET("", handle.List[
			controllers.ListAchievementsFilters,
			controllers.AchievementWithImage,
		](store.Achievements))
	}
}
