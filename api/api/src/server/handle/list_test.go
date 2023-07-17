package handle

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"

	"bitbucket.org/pensarmais/cycleforlisbon/src/server/controllers"
	"bitbucket.org/pensarmais/cycleforlisbon/src/util/httputil"
	"github.com/gin-gonic/gin"
)

func (m *MockController) List(params TestQuery, c *gin.Context) ([]TestType, error) {
	args := m.Called(params)
	return args.Get(0).([]TestType), args.Error(1)
}

// The `List` handler calls the `List` function of the controller with the
// correct parameters.
func (s *HandlersTestSuite) TestListHandler() {
	args := TestQuery{
		Pagination: controllers.Pagination{
			Limit:  5,
			Offset: 3,
		},
		Required: "abc",
	}

	s.controller.On("List", args).Return(make([]TestType, args.Limit), nil)

	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test?limit=5&offset=3&required=abc", nil)

	s.router.ServeHTTP(res, req)

	s.Equal(http.StatusOK, res.Result().StatusCode)

	resBody := []TestType{}
	if err := json.Unmarshal(res.Body.Bytes(), &resBody); err != nil {
		s.FailNow(err.Error())
	}

	expected := make([]TestType, args.Limit)
	s.Equal(expected, resBody)

	s.controller.AssertExpectations(s.T())
}

// The `List` handler validates the query parameters.
func (s *HandlersTestSuite) TestListHandlerInvalidQuery() {
	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	s.router.ServeHTTP(res, req)

	s.Equal(http.StatusBadRequest, res.Result().StatusCode)

	var resBody ErrorResponse
	if err := json.Unmarshal(res.Body.Bytes(), &resBody); err != nil {
		s.FailNow(err.Error())
	}

	expected := ErrorResponse{httputil.Error{
		Code:    "Bad Request",
		Message: "Key: 'TestQuery.Required' Error:Field validation for 'Required' failed on the 'required' tag",
	}}
	s.Equal(expected, resBody)
}

// The `List` handler returns errors from the controller.
func (s *HandlersTestSuite) TestListHandlerError() {
	args := TestQuery{
		Pagination: controllers.Pagination{
			Limit:  0,
			Offset: 0,
		},
		Required: "abc",
	}

	s.controller.On("List", args).Return(
		make([]TestType, args.Limit),
		errors.New("mm, don't like it"),
	)

	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test?required=abc", nil)

	s.router.ServeHTTP(res, req)

	s.Equal(http.StatusInternalServerError, res.Result().StatusCode)

	var resBody ErrorResponse
	if err := json.Unmarshal(res.Body.Bytes(), &resBody); err != nil {
		s.FailNow(err.Error())
	}

	expected := ErrorResponse{httputil.Error{
		Code:    "Internal Server Error",
		Message: "mm, don't like it"},
	}
	s.Equal(expected, resBody)

	s.controller.AssertExpectations(s.T())
}
