package handle

import (
	"net/http"

	"bitbucket.org/pensarmais/cycleforlisbon/src/util/httputil"
	"github.com/gin-gonic/gin"
)

type CreateController[K, T any] interface {
	Create(params K, c *gin.Context) (T, error)
}

func Create[K, T any](controller CreateController[K, T]) gin.HandlerFunc {
	return WrapCreate(controller.Create)
}

func WrapCreate[K, T any](create func(K, *gin.Context) (T, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		var params K
		if err := c.ShouldBindJSON(&params); err != nil {
			c.Error(httputil.NewError(httputil.BadRequest, err))
			return
		}

		result, err := create(params, c)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusCreated, result)
	}
}
