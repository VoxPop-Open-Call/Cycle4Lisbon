package route

import (
	"bitbucket.org/pensarmais/cycleforlisbon/src/database/models"
	"bitbucket.org/pensarmais/cycleforlisbon/src/server/controllers"
	"bitbucket.org/pensarmais/cycleforlisbon/src/server/handle"
	"github.com/gin-gonic/gin"
)

func Trips(
	router *gin.RouterGroup,
	auth gin.HandlerFunc,
	store *controllers.Store,
) {
	trips := router.Group("/trips", auth)
	{
		trips.GET("", handle.List[
			controllers.ListTripsFilters,
			models.Trip,
		](store.Trips))

		trips.GET("/:id", handle.Get[models.Trip](store.Trips))
		trips.GET("/:id/file", handle.Download(store.Trips))

		trips.POST("", handle.Upload[models.Trip](store.Trips))
	}
}
