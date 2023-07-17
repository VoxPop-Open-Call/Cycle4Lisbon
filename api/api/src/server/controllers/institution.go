package controllers

import (
	"errors"

	"bitbucket.org/pensarmais/cycleforlisbon/src/database/models"
	"bitbucket.org/pensarmais/cycleforlisbon/src/util/httputil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type institutionLogoPresigner interface {
	PresignGetInstitutionLogo(institutionID string) (string, string, error)
	PresignPutInstitutionLogo(institutionID string) (string, string, error)
	PresignDeleteInstitutionLogo(institutionID string) (string, string, error)
}

type InstitutionController struct {
	db        *gorm.DB
	acl       authorizer
	presigner institutionLogoPresigner
}

// Rules returns the acl for the institution controller.
func (InstitutionController) Rules() []rule {
	return []rule{
		{models.User{}, models.Institution{},
			"create,update,delete", func(ent, _ any) bool {
				return ent.(models.User).Admin
			},
		},
		{models.User{}, models.Institution{},
			"update-logo,delete-logo", func(ent, _ any) bool {
				return ent.(models.User).Admin
			},
		},
	}
}

type InstitutionWithImage struct {
	models.Institution
	PresignedLogoURL string `json:"presignedLogoURL,omitempty"`
}

// institutionWithImage enriches the given institution with a pre-signed URL to
// retrieve its logo.
func institutionWithImage(
	institution models.Institution,
	presigner institutionLogoPresigner,
) InstitutionWithImage {
	url, _, _ := presigner.PresignGetInstitutionLogo(institution.ID.String())
	return InstitutionWithImage{
		Institution:      institution,
		PresignedLogoURL: url,
	}
}

func institutionsWithImage(
	institutions []models.Institution,
	presigner institutionLogoPresigner,
) []InstitutionWithImage {
	res := make([]InstitutionWithImage, len(institutions))
	for i, institution := range institutions {
		res[i] = institutionWithImage(institution, presigner)
	}
	return res
}

type ListInstitutionsFilters struct {
	Pagination
	Sort
}

// List all institutions.
//
//	@Summary	List all institutions
//	@Tags		institutions
//	@Produce	json
//	@Security	OIDCToken
//	@Security	AuthHeader
//	@Param		filters		query		ListInstitutionsFilters	false	"Filters"
//	@Success	200			{array}		InstitutionWithImage
//	@Failure	400,401,500	{object}	middleware.ApiError
//	@Router		/institutions  [get]
func (c *InstitutionController) List(
	filters ListInstitutionsFilters,
	_ *gin.Context,
) ([]InstitutionWithImage, error) {
	var institutions []models.Institution
	err := c.db.
		Limit(filters.Limit).
		Offset(filters.Offset).
		Order(filters.OrderBy.ToSnakeCase()).
		Find(&institutions).Error

	return institutionsWithImage(institutions, c.presigner), err
}

// Get retrieves a institution.
//
//	@Summary	Retrieve a institution by ID
//	@Tags		institutions
//	@Produce	json
//	@Security	OIDCToken
//	@Security	AuthHeader
//	@Param		id				path		string	true	"Institution Id"	Format(UUID)
//	@Success	200				{object}	InstitutionWithImage
//	@Failure	400,401,404,500	{object}	middleware.ApiError
//	@Router		/institutions/{id} [get]
func (c *InstitutionController) Get(id string, ctx *gin.Context) (InstitutionWithImage, error) {
	var institution models.Institution
	err := c.db.First(&institution, "id = ?", id).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return InstitutionWithImage{}, resourceNotFoundErr("institution")
	}

	return institutionWithImage(institution, c.presigner), err
}

type CreateInstitutionParams struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description" binding:"required"`
}

// Create a institution.
//
//	@Summary	Create a new institution and return it
//	@Tags		institutions
//	@Produce	json
//	@Security	OIDCToken
//	@Security	AuthHeader
//	@Param		params		body		CreateInstitutionParams	true	"Params"
//	@Success	201			{object}	models.Institution
//	@Failure	400,401,500	{object}	middleware.ApiError
//	@Router		/institutions [post]
func (c *InstitutionController) Create(
	params CreateInstitutionParams,
	ctx *gin.Context,
) (models.Institution, error) {
	user, err := tokenUser(ctx, c.db)
	if err != nil {
		return models.Institution{}, err
	}

	institution := models.Institution{
		Name:        params.Name,
		Description: params.Description,
	}

	if ok := c.acl.Authorize(
		user, "create", institution,
	); !ok {
		return models.Institution{}, httputil.NewErrorMsg(
			httputil.AdminAccessRequired,
			httputil.AdminRequiredMessage,
		)
	}

	err = c.db.Create(&institution).Error

	return institution, err
}

// GetLogoURL generates a pre-signed url to retrieve the institution's logo.
//
//	@Summary		Generate a pre-signed url to retrieve the institution's logo
//	@Description	This method is only added for completion, since the methods to Retrieve and List
//	@Description	institutions already include this pre-signed URL in the response.
//	@Tags			institutions
//	@Produce		json
//	@Security		OIDCToken
//	@Security		AuthHeader
//	@Param			id				path		string	true	"Institution Id"	Format(UUID)
//	@Success		200				{object}	PresignedResponse
//	@Failure		400,401,404,500	{object}	middleware.ApiError
//	@Router			/institutions/{id}/logo-get-url [get]
func (c *InstitutionController) GetLogoURL(
	id string,
	_ *gin.Context,
) (PresignedResponse, error) {
	url, method, err := c.presigner.PresignGetInstitutionLogo(id)
	return PresignedResponse{url, method}, err
}

// PutImageURL generates a pre-signed url to update the institution's logo.
//
//	@Summary	Generate a pre-signed url to update the institution's logo
//	@Tags		institutions
//	@Produce	json
//	@Security	OIDCToken
//	@Security	AuthHeader
//	@Param		id					path		string	true	"Institution Id"	Format(UUID)
//	@Success	200					{object}	PresignedResponse
//	@Failure	400,401,403,404,500	{object}	middleware.ApiError
//	@Router		/institutions/{id}/logo-put-url [get]
func (c *InstitutionController) PutLogoURL(
	id string,
	ctx *gin.Context,
) (PresignedResponse, error) {
	user, err := tokenUser(ctx, c.db)
	if err != nil {
		return PresignedResponse{}, err
	}

	if ok := c.acl.Authorize(
		user, "update-logo", models.Institution{},
	); !ok {
		return PresignedResponse{}, httputil.NewErrorMsg(
			httputil.Forbidden,
			httputil.ForbiddenMessage,
		)
	}

	url, method, err := c.presigner.PresignPutInstitutionLogo(id)
	return PresignedResponse{url, method}, err
}

// DeleteImageURL generates a pre-signed url to delete the institution's logo.
//
//	@Summary	Generate a pre-signed url to delete the institution's logo
//	@Tags		institutions
//	@Produce	json
//	@Security	OIDCToken
//	@Security	AuthHeader
//	@Param		id					path		string	true	"Institution Id"	Format(UUID)
//	@Success	200					{object}	PresignedResponse
//	@Failure	400,401,403,404,500	{object}	middleware.ApiError
//	@Router		/institutions/{id}/logo-delete-url [get]
func (c *InstitutionController) DeleteLogoURL(
	id string,
	ctx *gin.Context,
) (PresignedResponse, error) {
	user, err := tokenUser(ctx, c.db)
	if err != nil {
		return PresignedResponse{}, err
	}

	if ok := c.acl.Authorize(
		user, "delete-logo", models.Institution{},
	); !ok {
		return PresignedResponse{}, httputil.NewErrorMsg(
			httputil.Forbidden,
			httputil.ForbiddenMessage,
		)
	}
	url, method, err := c.presigner.PresignDeleteInstitutionLogo(id)
	return PresignedResponse{url, method}, err
}

// Delete a institution.
//
//	@Summary	Delete a institution by Id
//	@Tags		institutions
//	@Security	OIDCToken
//	@Security	AuthHeader
//	@Param		id	path	string	true	"Institution Id"	Format(UUID)
//	@Success	204
//	@Failure	400,401,403,404,500	{object}	middleware.ApiError
//	@Router		/institutions/{id} [delete]
func (c *InstitutionController) Delete(id string, ctx *gin.Context) error {
	user, err := tokenUser(ctx, c.db)
	if err != nil {
		return err
	}

	uid, err := uuid.Parse(id)
	if err != nil {
		return httputil.NewError(httputil.BadRequest, err)
	}

	institution := models.Institution{
		BaseModel: models.BaseModel{ID: uid},
	}

	if ok := c.acl.Authorize(
		user, "delete", institution,
	); !ok {
		return httputil.NewErrorMsg(
			httputil.AdminAccessRequired,
			httputil.AdminRequiredMessage,
		)
	}

	result := c.db.Delete(&models.Institution{}, "id = ?", id)
	if result.RowsAffected == 0 {
		return resourceNotFoundErr("institution")
	}
	return result.Error
}
