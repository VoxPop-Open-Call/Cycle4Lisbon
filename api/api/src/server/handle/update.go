package handle

import (
	"net/http"

	"bitbucket.org/pensarmais/cycleforlisbon/src/util/httputil"
	"github.com/gin-gonic/gin"
)

type UpdateController[K, T any] interface {
	Update(id string, params K, c *gin.Context) (T, error)
}

func Update[K any, T any](controller UpdateController[K, T]) gin.HandlerFunc {
	return WrapUpdate(controller.Update)
}

func WrapUpdate[K, T any](update func(string, K, *gin.Context) (T, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := BindID(c)
		if err != nil {
			c.Error(err)
			return
		}

		var params K
		if err := c.ShouldBindJSON(&params); err != nil {
			c.Error(httputil.NewError(httputil.BadRequest, err))
			return
		}

		result, err := update(id, params, c)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, result)
	}
}
