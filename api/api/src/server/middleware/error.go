package middleware

import (
	"bitbucket.org/pensarmais/cycleforlisbon/src/util/httputil"
	"github.com/gin-gonic/gin"
)

// ApiError represents the format of error responses returned by the API.
type ApiError struct {
	httputil.Error `json:"error"`
}

// Error middleware formats the errors from the gin context, and serializes
// them into the response body.
//
// If the error is an httputil.Error it will be returned as is. Otherwise, an
// InternalServerError will be returned with the given error message.
func Error() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		ginErr := c.Errors.Last()
		if ginErr == nil {
			return
		}
		err := ginErr.Err

		customErr, ok := err.(httputil.Error)
		if !ok {
			customErr = httputil.NewError(
				httputil.InternalServerError,
				err,
			)
		}

		c.JSON(customErr.Status, ApiError{customErr})
	}
}
