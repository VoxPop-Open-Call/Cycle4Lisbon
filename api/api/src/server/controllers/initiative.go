package controllers

import (
	"errors"

	"bitbucket.org/pensarmais/cycleforlisbon/src/database/models"
	"bitbucket.org/pensarmais/cycleforlisbon/src/database/query"
	"bitbucket.org/pensarmais/cycleforlisbon/src/database/types"
	"bitbucket.org/pensarmais/cycleforlisbon/src/util/httputil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type initiativeImgPresigner interface {
	institutionLogoPresigner
	PresignGetInitiativeImg(initiativeID string) (string, string, error)
	PresignPutInitiativeImg(initiativeID string) (string, string, error)
	PresignDeleteInitiativeImg(initiativeID string) (string, string, error)
}

type InitiativeController struct {
	db        *gorm.DB
	acl       authorizer
	presigner initiativeImgPresigner
}

// Rules returns the acl for the initiative controller.
func (InitiativeController) Rules() []rule {
	return []rule{
		{models.User{}, models.Initiative{},
			"create,update,change-state,delete", func(ent, _ any) bool {
				return ent.(models.User).Admin
			}},
		{models.User{}, models.Initiative{},
			"update-img,delete-img", func(ent, _ any) bool {
				return ent.(models.User).Admin
			},
		},
	}
}

type InitiativeWithImage struct {
	models.Initiative
	PresignedImgURL      string                 `json:"presignedImageURL,omitempty"`
	InstitutionWithImage InstitutionWithImage   `json:"institution"`
	SponsorsWithImage    []InstitutionWithImage `json:"sponsors,omitempty"`
}

// initiativeWithImage enriches the given initiative with a pre-signed URL to
// retrieve its image.
//
// It also dives into the initiative and generates URLs for the institution and
// sponsors.
func initiativeWithImage(
	initiative models.Initiative,
	presigner initiativeImgPresigner,
) InitiativeWithImage {
	res := InitiativeWithImage{}
	res.Initiative = initiative
	res.InstitutionWithImage = institutionWithImage(res.Institution, presigner)
	res.SponsorsWithImage = institutionsWithImage(res.Sponsors, presigner)

	url, _, _ := presigner.PresignGetInitiativeImg(initiative.ID.String())
	res.PresignedImgURL = url
	return res
}

func initiativesWithImage(
	initiatives []models.Initiative,
	presigner initiativeImgPresigner,
) []InitiativeWithImage {
	res := make([]InitiativeWithImage, len(initiatives))
	for i, initiative := range initiatives {
		res[i] = initiativeWithImage(initiative, presigner)
	}
	return res
}

type ListInitiativesFilters struct {
	Pagination
	Sort
	// IncludeDisabled initiatives in the result. This value is ignored for
	// non-admin users, which only receive enabled initiatives.
	IncludeDisabled bool `form:"includeDisabled"`
}

// List all initiatives.
//
//	@Summary		List all initiatives
//	@Description	Only `enabled` initiatives will be returned for non-admin users.
//	@Tags			initiatives
//	@Produce		json
//	@Security		OIDCToken
//	@Security		AuthHeader
//	@Param			filters		query		ListInitiativesFilters	false	"Filters"
//	@Success		200			{array}		InitiativeWithImage
//	@Failure		400,401,500	{object}	middleware.ApiError
//	@Router			/initiatives  [get]
func (c *InitiativeController) List(
	filters ListInitiativesFilters,
	ctx *gin.Context,
) ([]InitiativeWithImage, error) {
	user, err := tokenUser(ctx, c.db)
	if err != nil {
		return nil, err
	}

	tx := query.Initiatives.WithAssociations(c.db)

	if !user.Admin || !filters.IncludeDisabled {
		tx = tx.Where("enabled = true")
	}

	var initiatives []models.Initiative
	err = tx.
		Limit(filters.Limit).
		Offset(filters.Offset).
		Order(filters.OrderBy.ToSnakeCase()).
		Find(&initiatives).Error

	return initiativesWithImage(initiatives, c.presigner), err
}

// Get retrieves an initiative.
//
//	@Summary	Retrieve an initiative by ID
//	@Tags		initiatives
//	@Produce	json
//	@Security	OIDCToken
//	@Security	AuthHeader
//	@Param		id				path		string	true	"Initiative Id"	Format(UUID)
//	@Success	200				{object}	InitiativeWithImage
//	@Failure	400,401,404,500	{object}	middleware.ApiError
//	@Router		/initiatives/{id} [get]
func (c *InitiativeController) Get(id string, ctx *gin.Context) (InitiativeWithImage, error) {
	var initiative models.Initiative
	err := query.Initiatives.WithAssociations(c.db).
		First(&initiative, "id = ?", id).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return InitiativeWithImage{}, resourceNotFoundErr("initiative")
	}

	return initiativeWithImage(initiative, c.presigner), err
}

type CreateInitiativeParams struct {
	Title         string               `json:"title" binding:"required"`
	Description   string               `json:"description" binding:"required"`
	Goal          uint32               `json:"goal" binding:"required"`
	EndDate       types.Date           `form:"endDate" binding:"required,datetime=2006-01-02" example:"2023-03-30"`
	InstitutionID uuid.UUID            `json:"institutionId" binding:"required"`
	Sponsors      []models.Institution `json:"sponsors,omitempty"`
	SDGs          []models.SDG         `json:"sdgs" binding:"omitempty,dive"`
}

// Create an initiative.
//
//	@Summary	Create a new initiative and return it
//	@Tags		initiatives
//	@Produce	json
//	@Security	OIDCToken
//	@Security	AuthHeader
//	@Param		params		body		CreateInitiativeParams	true	"Params"
//	@Success	201			{object}	models.Initiative
//	@Failure	400,401,500	{object}	middleware.ApiError
//	@Router		/initiatives [post]
func (c *InitiativeController) Create(
	params CreateInitiativeParams,
	ctx *gin.Context,
) (models.Initiative, error) {
	user, err := tokenUser(ctx, c.db)
	if err != nil {
		return models.Initiative{}, err
	}

	initiative := models.Initiative{
		Title:         params.Title,
		Description:   params.Description,
		Goal:          params.Goal,
		EndDate:       params.EndDate,
		InstitutionID: params.InstitutionID,
		Sponsors:      params.Sponsors,
		SDGs:          params.SDGs,
	}

	if ok := c.acl.Authorize(
		user, "create", initiative,
	); !ok {
		return models.Initiative{}, httputil.NewErrorMsg(
			httputil.AdminAccessRequired,
			httputil.AdminRequiredMessage,
		)
	}

	err = c.db.Create(&initiative).Error
	if err != nil {
		return models.Initiative{}, err
	}

	res, err := c.Get(initiative.ID.String(), ctx)
	return res.Initiative, err
}

func (c *InitiativeController) setInitiativeState(
	id string,
	enabled bool,
	ctx *gin.Context,
) (models.Initiative, error) {
	user, err := tokenUser(ctx, c.db)
	if err != nil {
		return models.Initiative{}, err
	}

	uid, err := uuid.Parse(id)
	if err != nil {
		return models.Initiative{},
			httputil.NewError(httputil.BadRequest, err)
	}

	initiative := models.Initiative{
		BaseModel: models.BaseModel{ID: uid},
		Enabled:   enabled,
	}

	if ok := c.acl.Authorize(
		user, "change-state", initiative,
	); !ok {
		return models.Initiative{}, httputil.NewErrorMsg(
			httputil.AdminAccessRequired,
			httputil.AdminRequiredMessage,
		)
	}

	result := c.db.
		Clauses(clause.Returning{}).
		Where("id = ?", id).
		Select("enabled").
		Updates(&initiative)

	if result.RowsAffected == 0 {
		return models.Initiative{},
			resourceNotFoundErr("initiative")
	}

	return initiative, result.Error
}

// Enable an initiative.
//
//	@Summary	Enable an initiative
//	@Tags		initiatives
//	@Produce	json
//	@Security	OIDCToken
//	@Security	AuthHeader
//	@Param		id					path		string	true	"Initiative Id"	Format(UUID)
//	@Success	200					{object}	models.Initiative
//	@Failure	400,401,403,404,500	{object}	middleware.ApiError
//	@Router		/initiatives/{id}/enable [put]
func (c *InitiativeController) Enable(
	id string,
	ctx *gin.Context,
) (models.Initiative, error) {
	return c.setInitiativeState(id, true, ctx)
}

// Disable an initiative.
//
//	@Summary	Disable an initiative
//	@Tags		initiatives
//	@Produce	json
//	@Security	OIDCToken
//	@Security	AuthHeader
//	@Param		id					path		string	true	"Initiative Id"	Format(UUID)
//	@Success	200					{object}	models.Initiative
//	@Failure	400,401,403,404,500	{object}	middleware.ApiError
//	@Router		/initiatives/{id}/disable [put]
func (c *InitiativeController) Disable(
	id string,
	ctx *gin.Context,
) (models.Initiative, error) {
	return c.setInitiativeState(id, false, ctx)
}

// GetImageURL generates a pre-signed url to retrieve the initiatives's banner image.
//
//	@Summary		Generate a pre-signed url to retrieve the initiative's banner image
//	@Description	This method is only added for completion, since the methods to Retrieve and List
//	@Description	initiatives already include this pre-signed URL in the response.
//	@Tags			initiatives
//	@Produce		json
//	@Security		OIDCToken
//	@Security		AuthHeader
//	@Param			id				path		string	true	"Initiative Id"	Format(UUID)
//	@Success		200				{object}	PresignedResponse
//	@Failure		400,401,404,500	{object}	middleware.ApiError
//	@Router			/initiatives/{id}/img-get-url [get]
func (c *InitiativeController) GetImageURL(
	id string,
	_ *gin.Context,
) (PresignedResponse, error) {
	url, method, err := c.presigner.PresignGetInitiativeImg(id)
	return PresignedResponse{url, method}, err
}

// PutImageURL generates a pre-signed url to update the initiative's banner image.
//
//	@Summary	Generate a pre-signed url to update the initiative's banner image
//	@Tags		initiatives
//	@Produce	json
//	@Security	OIDCToken
//	@Security	AuthHeader
//	@Param		id					path		string	true	"Initiative Id"	Format(UUID)
//	@Success	200					{object}	PresignedResponse
//	@Failure	400,401,403,404,500	{object}	middleware.ApiError
//	@Router		/initiatives/{id}/img-put-url [get]
func (c *InitiativeController) PutImageURL(
	id string,
	ctx *gin.Context,
) (PresignedResponse, error) {
	user, err := tokenUser(ctx, c.db)
	if err != nil {
		return PresignedResponse{}, err
	}

	if ok := c.acl.Authorize(
		user, "update-img", models.Initiative{},
	); !ok {
		return PresignedResponse{}, httputil.NewErrorMsg(
			httputil.Forbidden,
			httputil.ForbiddenMessage,
		)
	}

	url, method, err := c.presigner.PresignPutInitiativeImg(id)
	return PresignedResponse{url, method}, err
}

// DeleteImageURL generates a pre-signed url to delete the initiatives's banner image.
//
//	@Summary	Generate a pre-signed url to delete the initiatives's banner image
//	@Tags		initiatives
//	@Produce	json
//	@Security	OIDCToken
//	@Security	AuthHeader
//	@Param		id					path		string	true	"Initiative Id"	Format(UUID)
//	@Success	200					{object}	PresignedResponse
//	@Failure	400,401,403,404,500	{object}	middleware.ApiError
//	@Router		/initiatives/{id}/img-delete-url [get]
func (c *InitiativeController) DeleteImageURL(
	id string,
	ctx *gin.Context,
) (PresignedResponse, error) {
	user, err := tokenUser(ctx, c.db)
	if err != nil {
		return PresignedResponse{}, err
	}

	if ok := c.acl.Authorize(
		user, "delete-img", models.Initiative{},
	); !ok {
		return PresignedResponse{}, httputil.NewErrorMsg(
			httputil.Forbidden,
			httputil.ForbiddenMessage,
		)
	}
	url, method, err := c.presigner.PresignDeleteInitiativeImg(id)
	return PresignedResponse{url, method}, err
}

// Delete an initiative.
//
//	@Summary	Delete an initiative by Id
//	@Tags		initiatives
//	@Security	OIDCToken
//	@Security	AuthHeader
//	@Param		id	path	string	true	"Initiative Id"	Format(UUID)
//	@Success	204
//	@Failure	400,401,403,404,500	{object}	middleware.ApiError
//	@Router		/initiatives/{id} [delete]
func (c *InitiativeController) Delete(id string, ctx *gin.Context) error {
	user, err := tokenUser(ctx, c.db)
	if err != nil {
		return err
	}

	uid, err := uuid.Parse(id)
	if err != nil {
		return httputil.NewError(httputil.BadRequest, err)
	}

	initiative := models.Initiative{
		BaseModel: models.BaseModel{ID: uid},
	}

	if ok := c.acl.Authorize(
		user, "delete", initiative,
	); !ok {
		return httputil.NewErrorMsg(
			httputil.AdminAccessRequired,
			httputil.AdminRequiredMessage,
		)
	}

	result := c.db.Delete(&models.Initiative{}, "id = ?", id)
	if result.RowsAffected == 0 {
		return resourceNotFoundErr("initiative")
	}
	return result.Error
}
