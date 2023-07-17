package handle

import (
	"net/http"

	"bitbucket.org/pensarmais/cycleforlisbon/src/util/httputil"
	"github.com/gin-gonic/gin"
)

type ListController[K, T any] interface {
	List(filters K, c *gin.Context) ([]T, error)
}

func List[K, T any](controller ListController[K, T]) gin.HandlerFunc {
	return WrapList(controller.List)
}

func WrapList[K, T any](list func(K, *gin.Context) ([]T, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		var filters K

		if err := c.ShouldBindQuery(&filters); err != nil {
			c.Error(httputil.NewError(httputil.BadRequest, err))
			return
		}

		result, err := list(filters, c)
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}
