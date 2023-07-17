package route

import (
	"bitbucket.org/pensarmais/cycleforlisbon/src/database/models"
	"bitbucket.org/pensarmais/cycleforlisbon/src/server/controllers"
	"bitbucket.org/pensarmais/cycleforlisbon/src/server/handle"
	"github.com/gin-gonic/gin"
)

func POIs(
	router *gin.RouterGroup,
	auth gin.HandlerFunc,
	store *controllers.Store,
) {
	pois := router.Group("/pois", auth)
	{
		pois.GET("", handle.List[
			controllers.ListPOIsFilters,
			models.PointOfInterest,
		](store.POIs))

		pois.POST("", handle.WrapUpload(store.POIs.Import))
	}
}
