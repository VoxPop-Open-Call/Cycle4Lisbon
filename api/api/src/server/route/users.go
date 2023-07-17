package route

import (
	"bitbucket.org/pensarmais/cycleforlisbon/src/database/models"
	"bitbucket.org/pensarmais/cycleforlisbon/src/server/controllers"
	"bitbucket.org/pensarmais/cycleforlisbon/src/server/handle"
	"github.com/gin-gonic/gin"
)

func Users(
	router *gin.RouterGroup,
	auth gin.HandlerFunc,
	store *controllers.Store,
) {
	users := router.Group("/users")
	{
		users.POST("", handle.Create[
			controllers.CreateUserParams,
			models.User,
		](store.Users))

		private := users.Group("", auth)
		{
			private.GET("", handle.List[
				controllers.ListUsersFilters,
				models.User,
			](store.Users))

			private.GET("/current", handle.WrapRetrieve(store.Users.GetCurrent))

			private.GET("/achievements", handle.WrapRetrieve(store.Users.Achievements))

			private.GET("/:id", handle.Get[models.User](store.Users))
			private.GET("/:id/picture-get-url", handle.WrapGet(store.Users.GetPictureURL))
			private.GET("/:id/picture-put-url", handle.WrapGet(store.Users.PutPictureURL))
			private.GET("/:id/picture-delete-url", handle.WrapGet(store.Users.DeletePictureURL))

			private.PUT("/:id", handle.Update[
				controllers.UpdateUserParams,
				models.User,
			](store.Users))

			private.DELETE("/:id", handle.Delete(store.Users))
		}
	}
}
