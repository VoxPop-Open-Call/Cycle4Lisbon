package handle

import (
	"net/http"

	"bitbucket.org/pensarmais/cycleforlisbon/src/util/httputil"
	"github.com/gin-gonic/gin"
)

type UploadController[T any] interface {
	Upload(data []byte, c *gin.Context) (T, error)
}

func Upload[T any](controller UploadController[T]) gin.HandlerFunc {
	return WrapUpload(controller.Upload)
}

func WrapUpload[T any](
	upload func(data []byte, c *gin.Context) (T, error),
) gin.HandlerFunc {
	return func(c *gin.Context) {
		fileHeader, err := c.FormFile("file")
		if err != nil {
			c.Error(httputil.NewError(httputil.InvalidFile, err))
			return
		}

		file, err := fileHeader.Open()
		if err != nil {
			c.Error(httputil.NewError(httputil.InvalidFile, err))
			return
		}

		data := make([]byte, fileHeader.Size)
		n, err := file.Read(data)
		if err != nil || n != int(fileHeader.Size) {
			c.Error(err)
			return
		}

		result, err := upload(data, c)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, result)
	}
}
