package handle

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"

	"bitbucket.org/pensarmais/cycleforlisbon/src/util/httputil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (m *MockController) Get(id string, c *gin.Context) (TestType, error) {
	args := m.Called(id)
	result := TestType{
		ID: args.String(0),
	}
	return result, args.Error(1)
}

// The `Get` handler calls the `Get` function of the controller with the
// correct parameters.
func (s *HandlersTestSuite) TestGetHandler() {
	uid, _ := uuid.NewRandom()
	id := uid.String()
	expected := TestType{id}

	s.controller.On("Get", id).Return(id, nil)

	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test/"+id, nil)

	s.router.ServeHTTP(res, req)

	s.Equal(http.StatusOK, res.Result().StatusCode)

	var resBody TestType
	if err := json.Unmarshal(res.Body.Bytes(), &resBody); err != nil {
		s.FailNow(err.Error())
	}
	s.Equal(expected, resBody)

	s.controller.AssertExpectations(s.T())
}

// The `Get` handler asserts whether the `id` is a valid UUID.
func (s *HandlersTestSuite) TestGetHandlerIdValidation() {
	id := "not-a-valid-id"

	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test/"+id, nil)

	s.router.ServeHTTP(res, req)

	s.Equal(http.StatusBadRequest, res.Result().StatusCode)

	var resBody ErrorResponse
	if err := json.Unmarshal(res.Body.Bytes(), &resBody); err != nil {
		s.FailNow(err.Error())
	}

	expected := ErrorResponse{httputil.Error{
		Code:    "Invalid UUID",
		Message: "Key: 'uri.ID' Error:Field validation for 'ID' failed on the 'uuid' tag",
	}}
	s.Equal(expected, resBody)
}

// The `Get` handler returns errors from the controller.
func (s *HandlersTestSuite) TestGetHandlerError() {
	uid, _ := uuid.NewRandom()
	id := uid.String()

	s.controller.On("Get", id).Return(id, httputil.NewError(
		httputil.RecordNotFound,
		errors.New("cannot get what doesn't exist"),
	))

	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test/"+id, nil)

	s.router.ServeHTTP(res, req)

	s.Equal(http.StatusNotFound, res.Result().StatusCode)

	var resBody ErrorResponse
	if err := json.Unmarshal(res.Body.Bytes(), &resBody); err != nil {
		s.FailNow(err.Error())
	}

	expected := ErrorResponse{httputil.Error{
		Code:    "Record Not Found",
		Message: "cannot get what doesn't exist",
	}}
	s.Equal(expected, resBody)

	s.controller.AssertExpectations(s.T())
}
