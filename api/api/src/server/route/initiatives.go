package route

import (
	"bitbucket.org/pensarmais/cycleforlisbon/src/database/models"
	"bitbucket.org/pensarmais/cycleforlisbon/src/server/controllers"
	"bitbucket.org/pensarmais/cycleforlisbon/src/server/handle"
	"github.com/gin-gonic/gin"
)

func Initiatives(
	router *gin.RouterGroup,
	auth gin.HandlerFunc,
	store *controllers.Store,
) {
	initiatives := router.Group("/initiatives", auth)
	{
		initiatives.GET("", handle.List[
			controllers.ListInitiativesFilters,
			controllers.InitiativeWithImage,
		](store.Initiatives))

		initiatives.GET("/:id", handle.Get[controllers.InitiativeWithImage](store.Initiatives))

		initiatives.POST("", handle.Create[
			controllers.CreateInitiativeParams,
			models.Initiative,
		](store.Initiatives))

		initiatives.GET("/:id/img-get-url", handle.WrapGet(store.Initiatives.GetImageURL))
		initiatives.GET("/:id/img-put-url", handle.WrapGet(store.Initiatives.PutImageURL))
		initiatives.GET("/:id/img-delete-url", handle.WrapGet(store.Initiatives.DeleteImageURL))

		initiatives.PUT("/:id/enable", handle.WrapAction(store.Initiatives.Enable))
		initiatives.PUT("/:id/disable", handle.WrapAction(store.Initiatives.Disable))

		initiatives.DELETE("/:id", handle.Delete(store.Initiatives))
	}
}
