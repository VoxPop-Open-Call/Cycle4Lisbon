package route

import (
	"bitbucket.org/pensarmais/cycleforlisbon/src/server/controllers"
	"bitbucket.org/pensarmais/cycleforlisbon/src/server/handle"
	"github.com/gin-gonic/gin"
)

func Leaderboard(
	router *gin.RouterGroup,
	auth gin.HandlerFunc,
	store *controllers.Store,
) {
	leaderboard := router.Group("/leaderboard", auth)
	{
		leaderboard.GET("", handle.WrapRetrieve(store.Leaderboard.List))
	}
}
