package route

import (
	"bitbucket.org/pensarmais/cycleforlisbon/src/database/models"
	"bitbucket.org/pensarmais/cycleforlisbon/src/server/controllers"
	"bitbucket.org/pensarmais/cycleforlisbon/src/server/handle"
	"github.com/gin-gonic/gin"
)

func Languages(
	router *gin.RouterGroup,
	store *controllers.Store,
) {
	router.GET("/languages", handle.List[
		controllers.ListLanguageFilters,
		models.Language,
	](store.Languages))
}
