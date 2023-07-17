package route

import (
	"fmt"
	"net/http"

	"bitbucket.org/pensarmais/cycleforlisbon/src/server/controllers"
	"bitbucket.org/pensarmais/cycleforlisbon/src/server/handle"
	"bitbucket.org/pensarmais/cycleforlisbon/src/util/httputil"
	"github.com/gin-gonic/gin"
)

func Password(
	router *gin.RouterGroup,
	auth gin.HandlerFunc,
	store *controllers.Store,
) {
	password := router.Group("/password")
	{
		password.PUT("/reset", handle.WrapPut(store.Password.RequestReset))
		password.PUT("/confirm-reset", handle.WrapPut(store.Password.ConfirmReset))
		password.GET("/redirect-reset", handleRedirect)

		private := password.Group("", auth)
		{
			private.PUT("", handle.WrapPut(store.Password.Update))
		}
	}
}

type redirectResetParams struct {
	Email string `form:"email" binding:"required"`
	Code  string `form:"code" binding:"required"`
}

func handleRedirect(c *gin.Context) {
	var params redirectResetParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.Error(httputil.NewError(httputil.BadRequest, err))
		return
	}

	c.Redirect(http.StatusFound,
		fmt.Sprintf("cfl://password-reset?email=%s&code=%s",
			params.Email, params.Code))
}
