package controllers

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"

	"bitbucket.org/pensarmais/cycleforlisbon/src/database/models"
	"bitbucket.org/pensarmais/cycleforlisbon/src/util/httputil"
	"bitbucket.org/pensarmais/cycleforlisbon/src/util/maputil"
	"bitbucket.org/pensarmais/cycleforlisbon/src/util/stringutil"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type POIController struct {
	db  *gorm.DB
	acl authorizer
}

func (POIController) Rules() []rule {
	return []rule{
		{models.User{}, models.PointOfInterest{}, "import", func(ent, res any) bool {
			return ent.(models.User).Admin
		}},
	}
}

type ListPOIsFilters struct {
	MinLat float64 `form:"minLat" binding:"required"`
	MaxLat float64 `form:"maxLat" binding:"required"`
	MinLon float64 `form:"minLon" binding:"required"`
	MaxLon float64 `form:"maxLon" binding:"required"`
}

// List all points of interest.
//
//	@Summary	List all points of interest
//	@Tags		points of interest
//	@Produce	json
//	@Security	OIDCToken
//	@Security	AuthHeader
//	@Param		filters		query		ListPOIsFilters	false	"Filters"
//	@Success	200			{array}		models.PointOfInterest
//	@Failure	400,401,500	{object}	middleware.ApiError
//	@Router		/pois  [get]
func (c *POIController) List(
	filters ListPOIsFilters,
	ctx *gin.Context,
) ([]models.PointOfInterest, error) {
	var points []models.PointOfInterest
	err := c.db.
		Where("lat >= ?", filters.MinLat).
		Where("lat <= ?", filters.MaxLat).
		Where("lon >= ?", filters.MinLon).
		Where("lon <= ?", filters.MaxLon).
		Find(&points).Error

	return points, err
}

type ImportPOIsResponse struct {
	Success bool `json:"success"`
}

type ImportPOIsQuery struct {
	Type string `form:"type" binding:"required,oneof=gira"`
}

// getCol returns the index of the column with the given name, or an error if
// it doesn't exist.
func getCol(header []string, colName string) (int, error) {
	for i, h := range header {
		if h == colName {
			return i, nil
		}
	}
	return -1, fmt.Errorf("column '%s' not found", colName)
}

const (
	nameCol     = "desigcomercial"
	positionCol = "position"
)

// readCols maps the column names to their index.
func readCols(r *csv.Reader) (map[string]int, error) {
	cols := map[string]int{
		nameCol:     -1,
		positionCol: -1,
	}

	header, err := r.Read()
	if err == io.EOF {
		return nil, httputil.NewErrorMsg(
			httputil.ImportCSVMissingHeader,
			"expected column names, found end-of-file",
		)
	}
	if err != nil {
		return nil, err
	}

	for k := range cols {
		cols[k], err = getCol(header, k)
		if err != nil {
			return nil, httputil.NewError(
				httputil.ImportMissingColumn, err)
		}
	}

	return cols, nil
}

// Import a Points Of Interest file.
//
//	@Summary		Import a Points Of Interest file
//	@Description	Required columns of the CSV file are "name" and "position"
//	@Tags			points of interest
//	@Produce		json
//	@Security		OIDCToken
//	@Security		AuthHeader
//	@Param			params			query		ImportPOIsQuery	true	"Query"
//	@Param			file			formData	file			true	"Params"
//	@Success		200				{object}	ImportPOIsResponse
//	@Failure		400,401,403,500	{object}	middleware.ApiError
//	@Router			/pois [post]
func (c *POIController) Import(
	data []byte,
	ctx *gin.Context,
) (ImportPOIsResponse, error) {
	var query ImportPOIsQuery
	if err := ctx.ShouldBindQuery(&query); err != nil {
		return ImportPOIsResponse{}, httputil.NewError(
			httputil.InvalidUUID, err)
	}

	user, err := tokenUser(ctx, c.db)
	if err != nil {
		return ImportPOIsResponse{}, err
	}

	if !c.acl.Authorize(user, "import", models.PointOfInterest{}) {
		return ImportPOIsResponse{}, httputil.NewErrorMsg(
			httputil.AdminAccessRequired,
			httputil.AdminRequiredMessage,
		)
	}

	r := csv.NewReader(bytes.NewReader(data))

	cols, err := readCols(r)
	if err != nil {
		return ImportPOIsResponse{}, err
	}

	points := make(map[string]models.PointOfInterest)
	line := 2
	for ; ; line++ {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return ImportPOIsResponse{}, httputil.NewError(
				httputil.ImportReadError, err)
		}

		// Ignore repeated records in a file.
		_, exists := points[rec[cols[nameCol]]]
		if exists {
			continue
		}

		name := rec[cols[nameCol]]
		coords, err := stringutil.AllFloats(rec[cols[positionCol]])
		if err != nil || len(coords) != 2 {
			return ImportPOIsResponse{}, httputil.NewError(
				httputil.ImportInvalidValue,
				fmt.Errorf("invalid value on line %d: %v", line, err),
			)
		}

		points[name] = models.PointOfInterest{
			Name: name,
			Type: query.Type,
			Lat:  coords[1],
			Lon:  coords[0],
		}
	}

	return ImportPOIsResponse{true}, c.db.
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "name"},
				{Name: "type"},
			},
			UpdateAll: true,
		}).
		CreateInBatches(maputil.Array(points), 50).Error
}
