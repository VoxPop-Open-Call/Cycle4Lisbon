package controllers

import (
	"errors"
	"regexp"

	"bitbucket.org/pensarmais/cycleforlisbon/src/database/models"
	"bitbucket.org/pensarmais/cycleforlisbon/src/database/query"
	"bitbucket.org/pensarmais/cycleforlisbon/src/util/httputil"
	"bitbucket.org/pensarmais/cycleforlisbon/src/util/password"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserController struct {
	db            *gorm.DB
	acl           authorizer
	serverBaseURL string
	presigner     interface {
		PresignGetProfilePicture(userID string) (string, string, error)
		PresignPutProfilePicture(userID string) (string, string, error)
		PresignDeleteProfilePicture(userID string) (string, string, error)
	}
}

// Rules returns the acl for the user controller.
func (UserController) Rules() []rule {
	return []rule{
		{models.User{}, models.User{},
			"update,update-picture,delete-picture",
			func(ent, res any) bool {
				user := ent.(models.User)
				params := res.(models.User)
				return user.ID == params.ID
			}},
		{models.User{}, models.User{}, "delete", func(ent, res any) bool {
			user := ent.(models.User)
			params := res.(models.User)
			return user.ID == params.ID || user.Admin
		}},
	}
}

var duplicateEmailRegex = regexp.MustCompile(
	"duplicate key value violates unique constraint \"users_email_key\"",
)
var duplicateUsernameRegex = regexp.MustCompile(
	"duplicate key value violates unique constraint \"idx_users_username\"",
)

type ListUsersFilters struct {
	Pagination
	Sort
}

// Lists all users.
//
//	@Summary	List all users
//	@Tags		users
//	@Produce	json
//	@Security	OIDCToken
//	@Security	AuthHeader
//	@Param		filters		query		ListUsersFilters	false	"Filters"
//	@Success	200			{array}		models.User
//	@Failure	400,401,500	{object}	middleware.ApiError
//	@Router		/users  [get]
func (c *UserController) List(
	filters ListUsersFilters,
	_ *gin.Context,
) ([]models.User, error) {
	var users []models.User

	err := c.db.Model(&models.User{}).
		Joins("Initiative").
		Limit(filters.Limit).
		Offset(filters.Offset).
		Order(filters.OrderBy.ToSnakeCase()).
		Find(&users).Error

	return users, err
}

// Get retrieves a user.
//
//	@Summary	Retrieve a user by Id
//	@Tags		users
//	@Produce	json
//	@Security	OIDCToken
//	@Security	AuthHeader
//	@Param		id				path		string	true	"User Id"	Format(UUID)
//	@Success	200				{object}	models.User
//	@Failure	400,401,404,500	{object}	middleware.ApiError
//	@Router		/users/{id} [get]
func (c *UserController) Get(id string, ctx *gin.Context) (models.User, error) {
	var user models.User
	err := c.db.Model(&models.User{}).
		Joins("Initiative").
		First(&user, "users.id = ?", id).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.User{}, resourceNotFoundErr("user")
	}

	return user, err
}

// Get retrieves the logged-in user.
//
//	@Summary	Retrieve the currently logged-in user
//	@Tags		users
//	@Produce	json
//	@Security	OIDCToken
//	@Security	AuthHeader
//	@Success	200				{object}	models.User
//	@Failure	400,401,404,500	{object}	middleware.ApiError
//	@Router		/users/current [get]
func (c *UserController) GetCurrent(ctx *gin.Context) (models.User, error) {
	user, err := tokenUser(ctx, c.db)
	if err != nil {
		return models.User{}, err
	}

	return c.Get(user.ID.String(), ctx)
}

type CreateUserParams struct {
	// Subject is the `sub` field from the ID token claims.
	// Required when registering with a third party provider.
	// Subject and Password are mutually exclusive.
	Subject string `json:"subject" binding:"required_without=Password,excluded_with=Password"`

	// Password is required when the Subject is not provided, i.e. when
	// registering using email and password.
	// Subject and Password are mutually exclusive.
	Password string `json:"password" binding:"required_without=Subject,excluded_with=Subject,omitempty,min=8"`

	Email string `json:"email" binding:"required,email"`
	Name  string `json:"name"`
}

type UserAchievementWithImage struct {
	models.UserAchievement
	AchievementWithImage AchievementWithImage `json:"achievement"`
}

func userAchievementsWithImage(
	userAchievements []models.UserAchievement,
	serverBaseURL string,
) []UserAchievementWithImage {
	res := make([]UserAchievementWithImage, len(userAchievements))
	for i, userAchievement := range userAchievements {
		res[i] = UserAchievementWithImage{}
		res[i].UserAchievement = userAchievement
		res[i].AchievementWithImage = achievementWithImage(
			userAchievement.Achievement, serverBaseURL,
		)
	}
	return res

}

// Lists the current user's achievements.
//
//	@Summary	List the current user's achievements
//	@Tags		users, achievements
//	@Produce	json
//	@Security	OIDCToken
//	@Security	AuthHeader
//	@Success	200			{array}		UserAchievementWithImage
//	@Failure	400,401,500	{object}	middleware.ApiError
//	@Router		/users/achievements  [get]
func (c *UserController) Achievements(
	ctx *gin.Context,
) ([]UserAchievementWithImage, error) {
	user, err := tokenUser(ctx, c.db)
	if err != nil {
		return nil, err
	}

	achs, err := query.Achievements.List(user.ID.String(), c.db)
	if err != nil {
		return nil, err
	}

	return userAchievementsWithImage(achs, c.serverBaseURL), nil
}

// Creates a user.
//
//	@Summary	Create a new user and return it
//	@Tags		users
//	@Produce	json
//	@Param		params		body		CreateUserParams	true	"Params"
//	@Success	201			{object}	models.User
//	@Failure	400,401,500	{object}	middleware.ApiError
//	@Router		/users [post]
func (c *UserController) Create(
	params CreateUserParams,
	_ *gin.Context,
) (models.User, error) {
	var hash string
	if params.Password != "" {
		var err error
		hash, err = password.Hash(params.Password)
		if err != nil {
			return models.User{},
				httputil.NewError(httputil.BadRequest, err)
		}
	}

	uid, err := uuid.NewRandom()
	if err != nil {
		return models.User{},
			httputil.NewError(httputil.BadRequest, err)
	}

	sub := params.Subject
	if sub == "" {
		sub = uid.String()
	}

	user := models.User{
		BaseModel: models.BaseModel{
			ID: uid,
		},
		Subject:        sub,
		Email:          params.Email,
		HashedPassword: hash,
		Profile: models.Profile{
			Name: params.Name,
		},
	}

	err = c.db.Create(&user).Error

	if err != nil && duplicateEmailRegex.MatchString(err.Error()) {
		return models.User{}, httputil.NewErrorMsg(
			httputil.EmailAlreadyRegistered,
			"a user with the given email already exists",
		)
	}

	return user, err
}

type UpdateUserParams struct {
	models.Profile
	Email        string     `json:"email" binding:"omitempty,email"`
	InitiativeID *uuid.UUID `json:"initiativeId,omitempty"`
}

// Updates a user.
//
//	@Summary	Update a user by Id and return it
//	@Tags		users
//	@Produce	json
//	@Security	OIDCToken
//	@Security	AuthHeader
//	@Param		id					path		string				true	"User Id"	Format(UUID)
//	@Param		params				body		UpdateUserParams	true	"Params"
//	@Success	200					{object}	models.User
//	@Failure	400,401,403,404,500	{object}	middleware.ApiError
//	@Router		/users/{id} [put]
func (c *UserController) Update(
	id string,
	params UpdateUserParams,
	ctx *gin.Context,
) (models.User, error) {
	user, err := tokenUser(ctx, c.db)
	if err != nil {
		return models.User{}, err
	}

	userID, err := uuid.Parse(id)
	if err != nil {
		return models.User{},
			httputil.NewError(httputil.BadRequest, err)
	}

	profile := params.Profile
	userParams := models.User{
		BaseModel:    models.BaseModel{ID: userID},
		Email:        params.Email,
		Profile:      profile,
		InitiativeID: params.InitiativeID,
	}

	if ok := c.acl.Authorize(
		user, "update", userParams,
	); !ok {
		return models.User{}, httputil.NewErrorMsg(
			httputil.Forbidden,
			httputil.ForbiddenMessage,
		)
	}

	if err := c.db.Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&userParams).
			Omit(clause.Associations).
			Where("id = ?", id).
			Updates(&userParams)

		if err := res.Error; err != nil {
			if duplicateEmailRegex.MatchString(err.Error()) {
				return httputil.NewErrorMsg(
					httputil.EmailAlreadyRegistered,
					"a user with the given email already exists",
				)
			}

			if duplicateUsernameRegex.MatchString(err.Error()) {
				return httputil.NewErrorMsg(
					httputil.UsernameAlreadyRegistered,
					"a user with the given username already exists",
				)
			}

			return res.Error
		}
		if res.RowsAffected == 0 {
			return resourceNotFoundErr("user")
		}

		return tx.Model(&userParams).
			Update("initiative_id", userParams.InitiativeID).
			Error
	}); err != nil {
		return models.User{}, err
	}

	return c.Get(id, ctx)
}

// GetPictureURL generates a pre-signed url to retrieve the user's profile picture.
//
//	@Summary	Generate a pre-signed url to retrieve the user's profile picture
//	@Tags		users
//	@Produce	json
//	@Security	OIDCToken
//	@Security	AuthHeader
//	@Param		id				path		string	true	"User Id"	Format(UUID)
//	@Success	200				{object}	PresignedResponse
//	@Failure	400,401,404,500	{object}	middleware.ApiError
//	@Router		/users/{id}/picture-get-url [get]
func (c *UserController) GetPictureURL(
	id string,
	_ *gin.Context,
) (PresignedResponse, error) {
	url, method, err := c.presigner.PresignGetProfilePicture(id)
	return PresignedResponse{url, method}, err
}

// PutPictureURL generates a pre-signed url to update the user's profile picture.
//
//	@Summary	Generate a pre-signed url to update the user's profile picture
//	@Tags		users
//	@Produce	json
//	@Security	OIDCToken
//	@Security	AuthHeader
//	@Param		id					path		string	true	"User Id"	Format(UUID)
//	@Success	200					{object}	PresignedResponse
//	@Failure	400,401,403,404,500	{object}	middleware.ApiError
//	@Router		/users/{id}/picture-put-url [get]
func (c *UserController) PutPictureURL(
	id string,
	ctx *gin.Context,
) (PresignedResponse, error) {
	user, err := tokenUser(ctx, c.db)
	if err != nil {
		return PresignedResponse{}, err
	}

	userID, err := uuid.Parse(id)
	if err != nil {
		return PresignedResponse{},
			httputil.NewError(httputil.BadRequest, err)
	}

	if ok := c.acl.Authorize(
		user, "update-picture", models.User{
			BaseModel: models.BaseModel{ID: userID},
		},
	); !ok {
		return PresignedResponse{}, httputil.NewErrorMsg(
			httputil.Forbidden,
			httputil.ForbiddenMessage,
		)
	}

	url, method, err := c.presigner.PresignPutProfilePicture(id)
	return PresignedResponse{url, method}, err
}

// DeletePictureURL generates a pre-signed url to delete the user's profile
// picture.
//
//	@Summary	Generate a pre-signed url to delete the user's profile picture
//	@Tags		users
//	@Produce	json
//	@Security	OIDCToken
//	@Security	AuthHeader
//	@Param		id					path		string	true	"User Id"	Format(UUID)
//	@Success	200					{object}	PresignedResponse
//	@Failure	400,401,403,404,500	{object}	middleware.ApiError
//	@Router		/users/{id}/picture-delete-url [get]
func (c *UserController) DeletePictureURL(
	id string,
	ctx *gin.Context,
) (PresignedResponse, error) {
	user, err := tokenUser(ctx, c.db)
	if err != nil {
		return PresignedResponse{}, err
	}

	userID, err := uuid.Parse(id)
	if err != nil {
		return PresignedResponse{},
			httputil.NewError(httputil.BadRequest, err)
	}

	if ok := c.acl.Authorize(
		user, "delete-picture", models.User{
			BaseModel: models.BaseModel{ID: userID},
		},
	); !ok {
		return PresignedResponse{}, httputil.NewErrorMsg(
			httputil.Forbidden,
			httputil.ForbiddenMessage,
		)
	}
	url, method, err := c.presigner.PresignDeleteProfilePicture(id)
	return PresignedResponse{url, method}, err
}

// Deletes a user.
//
//	@Summary	Delete a user by Id
//	@Tags		users
//	@Security	OIDCToken
//	@Security	AuthHeader
//	@Param		id	path	string	true	"User Id"	Format(UUID)
//	@Success	204
//	@Failure	400,401,403,404,500	{object}	middleware.ApiError
//	@Router		/users/{id} [delete]
func (c *UserController) Delete(id string, ctx *gin.Context) error {
	user, err := tokenUser(ctx, c.db)
	if err != nil {
		return err
	}

	userID, err := uuid.Parse(id)
	if err != nil {
		return httputil.NewError(httputil.BadRequest, err)
	}

	if ok := c.acl.Authorize(
		user, "delete", models.User{
			BaseModel: models.BaseModel{ID: userID},
		},
	); !ok {
		return httputil.NewErrorMsg(
			httputil.Forbidden,
			httputil.ForbiddenMessage,
		)
	}

	result := c.db.Delete(&models.User{}, "id = ?", id)
	if result.RowsAffected == 0 {
		return resourceNotFoundErr("user")
	}
	return result.Error
}
