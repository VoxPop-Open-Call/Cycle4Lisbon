package handle

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type DeleteController interface {
	Delete(id string, c *gin.Context) error
}

func Delete(controller DeleteController) gin.HandlerFunc {
	return WrapDelete(controller.Delete)
}

func WrapDelete(delete func(string, *gin.Context) error) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := BindID(c)
		if err != nil {
			c.Error(err)
			return
		}

		if err := delete(id, c); err != nil {
			c.Error(err)
			return
		}

		c.Status(http.StatusNoContent)
	}
}
