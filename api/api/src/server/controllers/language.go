package controllers

import (
	"bitbucket.org/pensarmais/cycleforlisbon/src/database/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type LanguageController struct {
	db *gorm.DB
}

type ListLanguageFilters struct {
	Pagination

	// OrderBy specifies the sorting order of the records returned by List
	// methods.
	//
	// Don't embed controllers.Sort, because the default value of `id asc` will
	// cause an error, since the Language model doesn't have that column.
	OrderBy orderBy `form:"orderBy,default=code asc" example:"foo,bar asc" default:"code asc"`
}

// Lists all languages.
//
//	@Summary	List all lanugages
//	@Tags		languages
//	@Produce	json
//	@Security	OIDCToken
//	@Security	AuthHeader
//	@Param		filters	query		ListLanguageFilters	false	"Filters"
//	@Success	200		{array}		models.Language
//	@Failure	400,500	{object}	middleware.ApiError
//	@Router		/languages  [get]
func (c *LanguageController) List(
	filters ListLanguageFilters,
	_ *gin.Context,
) ([]models.Language, error) {
	var langs []models.Language
	err := c.db.
		Limit(filters.Limit).
		Offset(filters.Offset).
		Order(filters.OrderBy.ToSnakeCase()).
		Find(&langs).Error

	return langs, err
}
