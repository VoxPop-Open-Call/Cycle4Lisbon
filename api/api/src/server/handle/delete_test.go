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

func (m *MockController) Delete(id string, _ *gin.Context) error {
	args := m.Called(id)
	return args.Error(0)
}

// The `Delete` handler calls the `Delete` function of the controller with the
// correct parameters.
func (s *HandlersTestSuite) TestDeleteHandler() {
	uid, _ := uuid.NewRandom()
	id := uid.String()

	s.controller.On("Delete", id).Return(nil)

	res := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/test/"+id, nil)

	s.router.ServeHTTP(res, req)

	s.Equal(http.StatusNoContent, res.Result().StatusCode)
	s.controller.AssertExpectations(s.T())
}

// The `Delete` handler asserts whether the `id` is a valid UUID.
func (s *HandlersTestSuite) TestDeleteHandlerIdValidation() {
	id := "abcde-132465"

	res := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/test/"+id, nil)

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

// The `Delete` handler returns errors from the controller.
func (s *HandlersTestSuite) TestDeleteHandlerError() {
	uid, _ := uuid.NewRandom()
	id := uid.String()

	s.controller.On("Delete", id).Return(httputil.NewError(
		httputil.RecordNotFound,
		errors.New("cannot delete what doesn't exist"),
	))

	res := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/test/"+id, nil)

	s.router.ServeHTTP(res, req)

	s.Equal(http.StatusNotFound, res.Result().StatusCode)

	var resBody ErrorResponse
	if err := json.Unmarshal(res.Body.Bytes(), &resBody); err != nil {
		s.FailNow(err.Error())
	}

	expected := ErrorResponse{httputil.Error{
		Code:    "Record Not Found",
		Message: "cannot delete what doesn't exist",
	}}
	s.Equal(expected, resBody)

	s.controller.AssertExpectations(s.T())
}
