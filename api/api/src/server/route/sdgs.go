package route

import (
	"bitbucket.org/pensarmais/cycleforlisbon/src/server/controllers"
	"bitbucket.org/pensarmais/cycleforlisbon/src/server/handle"
	"github.com/gin-gonic/gin"
)

func SDGs(
	router *gin.RouterGroup,
	auth gin.HandlerFunc,
	store *controllers.Store,
) {
	sdgs := router.Group("/sdgs", auth)
	{
		sdgs.GET("", handle.List[
			controllers.ListSDGFilters,
			controllers.SDGWithImage,
		](store.SDGs))
	}
}
