package handle

import (
	"testing"

	"bitbucket.org/pensarmais/cycleforlisbon/src/server/controllers"
	"bitbucket.org/pensarmais/cycleforlisbon/src/server/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// Setup a test suite for all handlers in this package, providing a mock router
// and controllers.

type HandlersTestSuite struct {
	suite.Suite
	router     *gin.Engine
	controller *MockController
}

type MockController struct {
	mock.Mock
}

type TestType struct {
	ID string `json:"id" binding:"uuid,required"`
}

type ErrorResponse middleware.ApiError

type TestQuery struct {
	controllers.Pagination
	Required string `form:"required" binding:"required"`
}

func TestHandlers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(middleware.Error())
	controller := &MockController{}

	router.GET("/test", List[TestQuery, TestType](controller))
	router.GET("/test/:id", Get[TestType](controller))
	router.POST("/test", Create[TestType, TestType](controller))
	router.PUT("/test/:id", Update[TestType, TestType](controller))
	router.DELETE("/test/:id", Delete(controller))

	suite.Run(t, &HandlersTestSuite{
		router:     router,
		controller: controller,
	})
}
