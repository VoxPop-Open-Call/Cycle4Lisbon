package controllers

import (
	"bitbucket.org/pensarmais/cycleforlisbon/src/database/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SDGController struct {
	db            *gorm.DB
	serverBaseURL string
}

type ListSDGFilters struct {
	Pagination

	// OrderBy specifies the sorting order of the records returned by List
	// methods.
	//
	// Don't embed controllers.Sort, because the default value of `id asc` will
	// cause an error, since the SDG model doesn't have that column.
	OrderBy orderBy `form:"orderBy,default=code asc" example:"foo,bar asc" default:"code asc"`
}

type SDGWithImage struct {
	models.SDG
	Image string `json:"image" gorm:"not null"`
}

func (c *SDGController) withImage(sdg models.SDG) SDGWithImage {
	return SDGWithImage{
		SDG:   sdg,
		Image: c.serverBaseURL + sdg.ImageURI,
	}
}

func (c *SDGController) allWithImage(sdgs []models.SDG) []SDGWithImage {
	res := make([]SDGWithImage, len(sdgs))
	for i, sdg := range sdgs {
		res[i] = c.withImage(sdg)
	}
	return res
}

// Lists all SDGs.
//
//	@Summary	List all SDGs
//	@Tags		sdgs
//	@Produce	json
//	@Security	OIDCToken
//	@Security	AuthHeader
//	@Param		filters	query		ListSDGFilters	false	"Filters"
//	@Success	200		{array}		SDGWithImage
//	@Failure	400,500	{object}	middleware.ApiError
//	@Router		/sdgs  [get]
func (c *SDGController) List(
	filters ListSDGFilters,
	_ *gin.Context,
) ([]SDGWithImage, error) {
	var sdgs []models.SDG
	err := c.db.
		Limit(filters.Limit).
		Offset(filters.Offset).
		Order(filters.OrderBy.ToSnakeCase()).
		Find(&sdgs).Error

	return c.allWithImage(sdgs), err
}
