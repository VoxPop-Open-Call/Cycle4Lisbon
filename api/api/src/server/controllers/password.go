package controllers

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"bitbucket.org/pensarmais/cycleforlisbon/src/database/models"
	"bitbucket.org/pensarmais/cycleforlisbon/src/database/query"
	"bitbucket.org/pensarmais/cycleforlisbon/src/util/httputil"
	"bitbucket.org/pensarmais/cycleforlisbon/src/util/password"
	"bitbucket.org/pensarmais/cycleforlisbon/src/util/random"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	// How long password reset codes are valid for.
	resetCodeLifetime = time.Hour
)

type PasswordController struct {
	db              *gorm.DB
	passwordEmailer interface {
		SendPasswordResetEmail(email, code string) error
		SendPasswordChangedEmail(email string) error
	}
}

type UpdatePasswordParams struct {
	Old string `json:"old" binding:"required"`
	New string `json:"new" binding:"required,min=8"`
}

// Updates a user's password.
//
//	@Summary	Update a user's password
//	@Tags		password
//	@Produce	json
//	@Security	OIDCToken
//	@Security	AuthHeader
//	@Param		params	body	UpdatePasswordParams	true	"Params"
//	@Success	204
//	@Failure	400,401,403,404,500	{object}	middleware.ApiError
//	@Router		/password [put]
func (c *PasswordController) Update(
	params UpdatePasswordParams,
	ctx *gin.Context,
) (int, error) {
	user, err := tokenUser(ctx, c.db)
	if err != nil {
		return 0, err
	}

	if !password.Check(params.Old, user.HashedPassword) {
		return 0, httputil.NewErrorMsg(httputil.IncorrectPassword,
			"the user's current password doesn't match the one provided")
	}

	hash, err := password.Hash(params.New)
	if err != nil {
		return 0, httputil.NewError(httputil.BadRequest, err)
	}

	user.HashedPassword = hash
	err = c.db.Save(&user).Error

	return http.StatusNoContent, err
}

type RequestPasswordResetParams struct {
	Email string `json:"email" binding:"required,email"`
}

// RequestReset initializes the password reset flow.
//
//	@Summary		Initialize the password reset flow
//	@Description	An email is sent to the provided address, iff the email is actually registered in the database.
//	@Description	No error is returned in case the email doesn't exist, to avoid leaking user data.
//	@Tags			password
//	@Produce		json
//	@Param			params	body	RequestPasswordResetParams	true	"Params"
//	@Success		202
//	@Failure		400,500	{object}	middleware.ApiError
//	@Router			/password/reset [put]
func (c *PasswordController) RequestReset(
	params RequestPasswordResetParams,
	_ *gin.Context,
) (int, error) {
	var err error

	if _, err = query.Users.ByEmail(params.Email, c.db); err != nil {
		log.Printf("refusing to send password reset email to %s: %v",
			params.Email, err)
		return http.StatusAccepted, nil
	}

	code := random.AlphanumericString(32)

	if err = c.db.Create(&models.PasswordResetCode{
		Code:      code,
		Email:     params.Email,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(resetCodeLifetime),
	}).Error; err != nil {
		return 0, fmt.Errorf("failed to store reset code in db: %v", err)
	}

	if err = c.passwordEmailer.
		SendPasswordResetEmail(params.Email, code); err != nil {
		return 0, fmt.Errorf("failed to send password reset email to %s: %v",
			params.Email, err)
	}

	return http.StatusAccepted, nil
}

type ConfirmPasswordResetParams struct {
	// Code provided to the user via email.
	Code string `json:"code" binding:"required"`
	New  string `json:"new" binding:"required,min=8"`
}

// ConfirmReset confirms a password reset and updates the password.
//
//	@Summary	Confirm a password reset and update the password
//	@Tags		password
//	@Produce	json
//	@Param		params	body	ConfirmPasswordResetParams	true	"Params"
//	@Success	204
//	@Failure	400,404,500	{object}	middleware.ApiError
//	@Router		/password/confirm-reset [put]
func (c *PasswordController) ConfirmReset(
	params ConfirmPasswordResetParams,
	_ *gin.Context,
) (int, error) {
	if err := c.db.Transaction(func(tx *gorm.DB) error {
		var record models.PasswordResetCode
		if err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&record, "code = ?", params.Code).
			Error; err != nil {
			return resourceNotFoundErr("reset code")
		}

		if record.IsExpired() {
			return httputil.NewErrorMsg(
				httputil.PasswordResetCodeExpired,
				"the password reset code is no longer valid",
			)
		}

		if record.Used {
			return httputil.NewErrorMsg(
				httputil.PasswordResetCodeUsed,
				"the password reset code has already been used",
			)
		}

		hash, err := password.Hash(params.New)
		if err != nil {
			return httputil.NewError(httputil.BadRequest, err)
		}

		res := tx.Model(&models.User{}).
			Where("email = ?", record.Email).
			Update("hashed_password", hash)
		if res.Error != nil {
			return fmt.Errorf("failed to update password: %v", err)
		}
		if res.RowsAffected == 0 {
			return resourceNotFoundErr("user")
		}

		record.Used = true
		if err = tx.Save(record).Error; err != nil {
			return fmt.Errorf("failed to flag record as used: %v", err)
		}

		if err = c.passwordEmailer.
			SendPasswordChangedEmail(record.Email); err != nil {
			log.Printf("failed to send password changed email to %s: %v",
				record.Email, err)
		}

		return nil
	}); err != nil {
		return 0, err
	}

	return http.StatusNoContent, nil
}
