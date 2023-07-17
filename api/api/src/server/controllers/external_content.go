package controllers

import (
	"bitbucket.org/pensarmais/cycleforlisbon/src/database/models"
	"bitbucket.org/pensarmais/cycleforlisbon/src/util/httputil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ExternalContentController struct {
	db  *gorm.DB
	acl authorizer
}

// Rules returns the acl for the external content controller.
func (ExternalContentController) Rules() []rule {
	return []rule{
		{models.User{}, models.ExternalContent{}, "change-state", func(ent, _ any) bool {
			user := ent.(models.User)
			return user.Admin
		}},
	}
}

type ListExternalContentFilters struct {
	Pagination
	Sort
	// Type of content to return.
	Type string `form:"type" binding:"required,oneof=event news" example:"event"`
	// State of the external content entries to list. This value is ignored
	// for non-admin users.
	// By default, all entries are returned except those `rejected`.
	State string `form:"state" binding:"omitempty,oneof=pending approved rejected" example:"approved"`
}

// List all external content entries.
//
//	@Summary		List all external content entries
//	@Description	Only `approved` content will be returned for non-admin users.
//	@Tags			external content
//	@Produce		json
//	@Security		OIDCToken
//	@Security		AuthHeader
//	@Param			filters		query		ListExternalContentFilters	false	"Filters"
//	@Success		200			{array}		models.ExternalContent
//	@Failure		400,401,500	{object}	middleware.ApiError
//	@Router			/external  [get]
func (c *ExternalContentController) List(
	filters ListExternalContentFilters,
	ctx *gin.Context,
) ([]models.ExternalContent, error) {
	user, err := tokenUser(ctx, c.db)
	if err != nil {
		return nil, err
	}

	tx := c.db.Model(&models.ExternalContent{}).
		Joins("Language")

	if !user.Admin {
		tx = tx.Where("state = ?", "approved")
	} else if filters.State != "" {
		tx = tx.Where("state = ?", filters.State)
	} else {
		tx = tx.Where("state != ?", "rejected")
	}

	var entries []models.ExternalContent
	err = tx.
		Where("type = ?", filters.Type).
		Limit(filters.Limit).
		Offset(filters.Offset).
		Order(filters.OrderBy.ToSnakeCase()).
		Find(&entries).Error

	return entries, err
}

func (c *ExternalContentController) setExternalContentState(
	id string,
	state models.ExternalContentState,
	ctx *gin.Context,
) (models.ExternalContent, error) {
	user, err := tokenUser(ctx, c.db)
	if err != nil {
		return models.ExternalContent{}, err
	}

	entryId, err := uuid.Parse(id)
	if err != nil {
		return models.ExternalContent{},
			httputil.NewError(httputil.BadRequest, err)
	}

	entry := models.ExternalContent{
		BaseModel: models.BaseModel{ID: entryId},
		State:     state,
	}

	if ok := c.acl.Authorize(
		user, "change-state", entry,
	); !ok {
		return models.ExternalContent{}, httputil.NewErrorMsg(
			httputil.AdminAccessRequired,
			httputil.AdminRequiredMessage,
		)
	}

	result := c.db.
		Clauses(clause.Returning{}).
		Where("id = ?", id).
		Updates(&entry)

	if result.RowsAffected == 0 {
		return models.ExternalContent{},
			resourceNotFoundErr("external content entry")
	}

	return entry, result.Error
}

// Approves an external content entry.
//
//	@Summary	Approve an external content entry by Id
//	@Tags		external content
//	@Produce	json
//	@Security	OIDCToken
//	@Security	AuthHeader
//	@Param		id					path		string	true	"ExternalContent Id"	Format(UUID)
//	@Success	200					{object}	models.ExternalContent
//	@Failure	400,401,403,404,500	{object}	middleware.ApiError
//	@Router		/external/{id}/approve [put]
func (c *ExternalContentController) Approve(
	id string,
	ctx *gin.Context,
) (models.ExternalContent, error) {
	return c.setExternalContentState(id, "approved", ctx)
}

// Rejects an external content entry.
//
//	@Summary	Reject an external content entry by Id
//	@Tags		external content
//	@Produce	json
//	@Security	OIDCToken
//	@Security	AuthHeader
//	@Param		id					path		string	true	"ExternalContent Id"	Format(UUID)
//	@Success	200					{object}	models.ExternalContent
//	@Failure	400,401,403,404,500	{object}	middleware.ApiError
//	@Router		/external/{id}/reject [put]
func (c *ExternalContentController) Reject(
	id string,
	ctx *gin.Context,
) (models.ExternalContent, error) {
	return c.setExternalContentState(id, "rejected", ctx)
}
