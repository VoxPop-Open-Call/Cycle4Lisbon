package handle

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type DownloadController interface {
	Download(id string, c *gin.Context) (data []byte, contentType string, err error)
}

func Download(controller DownloadController) gin.HandlerFunc {
	return WrapDownload(controller.Download)
}

func WrapDownload(
	download func(id string, c *gin.Context) ([]byte, string, error),
) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := BindID(c)
		if err != nil {
			c.Error(err)
			return
		}

		data, contentType, err := download(id, c)
		if err != nil {
			c.Error(err)
			return
		}

		c.Data(http.StatusOK, contentType, data)
	}
}
