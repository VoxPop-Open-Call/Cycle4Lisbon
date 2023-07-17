package middleware

import (
	"net/http"

	"bitbucket.org/pensarmais/cycleforlisbon/src/util/httputil"
	"github.com/gin-gonic/gin"
)

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		httputil.WriteCorsHeaders(c.Writer.Header())

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	}
}
