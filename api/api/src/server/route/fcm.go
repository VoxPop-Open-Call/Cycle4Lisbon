package route

import (
	"bitbucket.org/pensarmais/cycleforlisbon/src/server/controllers"
	"bitbucket.org/pensarmais/cycleforlisbon/src/server/handle"
	"github.com/gin-gonic/gin"
)

func FCM(
	router *gin.RouterGroup,
	auth gin.HandlerFunc,
	store *controllers.Store,
) {
	fcm := router.Group("/fcm", auth)
	{
		fcm.POST("/register", handle.WrapCreate(store.FCMTokens.Register))
	}
}
