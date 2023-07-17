package route

import (
	"bitbucket.org/pensarmais/cycleforlisbon/src/server/controllers"
	"bitbucket.org/pensarmais/cycleforlisbon/src/server/handle"
	"github.com/gin-gonic/gin"
)

func Metrics(
	router *gin.RouterGroup,
	auth gin.HandlerFunc,
	store *controllers.Store,
) {
	metrics := router.Group("/metrics", auth)
	{
		metrics.GET("", handle.WrapRetrieve(store.Metrics.Get))
	}
}
