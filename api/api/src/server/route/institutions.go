package route

import (
	"bitbucket.org/pensarmais/cycleforlisbon/src/database/models"
	"bitbucket.org/pensarmais/cycleforlisbon/src/server/controllers"
	"bitbucket.org/pensarmais/cycleforlisbon/src/server/handle"
	"github.com/gin-gonic/gin"
)

func Institutions(
	router *gin.RouterGroup,
	auth gin.HandlerFunc,
	store *controllers.Store,
) {
	institutions := router.Group("/institutions", auth)
	{
		institutions.GET("", handle.List[
			controllers.ListInstitutionsFilters,
			controllers.InstitutionWithImage,
		](store.Institutions))

		institutions.GET("/:id", handle.Get[controllers.InstitutionWithImage](store.Institutions))

		institutions.POST("", handle.Create[
			controllers.CreateInstitutionParams,
			models.Institution,
		](store.Institutions))

		institutions.GET("/:id/logo-get-url", handle.WrapGet(store.Institutions.GetLogoURL))
		institutions.GET("/:id/logo-put-url", handle.WrapGet(store.Institutions.PutLogoURL))
		institutions.GET("/:id/logo-delete-url", handle.WrapGet(store.Institutions.DeleteLogoURL))

		institutions.DELETE("/:id", handle.Delete(store.Institutions))
	}
}
