package handle

import (
	"net/http"

	"bitbucket.org/pensarmais/cycleforlisbon/src/util/httputil"
	"github.com/gin-gonic/gin"
)

type uri struct {
	ID string `uri:"id" binding:"required,uuid"`
}

func BindID(c *gin.Context) (string, error) {
	var params uri
	if err := c.ShouldBindUri(&params); err != nil {
		return "", httputil.NewError(
			httputil.InvalidUUID,
			err,
		)
	}

	return params.ID, nil
}

// WrapAction wraps a handler that takes the ID of a resource as an argument
// and returns a value of type T.
func WrapAction[T any](
	act func(id string, c *gin.Context) (T, error),
) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := BindID(c)
		if err != nil {
			c.Error(err)
			return
		}

		result, err := act(id, c)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

// WrapRetrieve wraps a handler that takes no arguments and returns a value of
// type T.
func WrapRetrieve[T any](
	retrieve func(c *gin.Context) (T, error),
) gin.HandlerFunc {
	return func(c *gin.Context) {
		result, err := retrieve(c)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

// WrapPut wraps a handler that takes an argument of type T and returns no
// values.
// The response status code is defined in the handler.
func WrapPut[T any](
	put func(params T, c *gin.Context) (int, error),
) gin.HandlerFunc {
	return func(c *gin.Context) {
		var params T
		if err := c.ShouldBindJSON(&params); err != nil {
			c.Error(httputil.NewError(httputil.BadRequest, err))
			return
		}

		status, err := put(params, c)
		if err != nil {
			c.Error(err)
			return
		}

		c.Status(status)
	}
}
