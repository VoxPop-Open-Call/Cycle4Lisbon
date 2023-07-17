package handle

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type GetController[T any] interface {
	Get(id string, c *gin.Context) (T, error)
}

func Get[T any](controller GetController[T]) gin.HandlerFunc {
	return WrapGet(controller.Get)
}

func WrapGet[T any](get func(string, *gin.Context) (T, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := BindID(c)
		if err != nil {
			c.Error(err)
			return
		}

		result, err := get(id, c)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, result)
	}
}
