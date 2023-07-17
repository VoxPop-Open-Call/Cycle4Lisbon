package handle

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"

	"bitbucket.org/pensarmais/cycleforlisbon/src/util/httputil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (m *MockController) Update(
	id string,
	params TestType,
	c *gin.Context,
) (TestType, error) {
	args := m.Called(id)
	return args.Get(0).(TestType), args.Error(1)
}

// The `Update` handler calls the `Update` function of the controller with the correct parameters.
func (s *HandlersTestSuite) TestUpdateHandler() {
	uid, _ := uuid.NewRandom()
	id := uid.String()
	expected := TestType{id}

	s.controller.On("Update", id).Return(expected, nil)

	data, err := json.Marshal(expected)
	if err != nil {
		s.FailNow("Unable to marshal request body")
	}
	res := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/test/"+id, bytes.NewReader(data))

	s.router.ServeHTTP(res, req)

	s.Equal(http.StatusOK, res.Result().StatusCode)

	var resBody TestType
	if err := json.Unmarshal(res.Body.Bytes(), &resBody); err != nil {
		s.FailNow(err.Error())
	}
	s.Equal(expected, resBody)

	s.controller.AssertExpectations(s.T())
}

// The `Update` handler asserts whether the `id` is a valid UUID.
func (s *HandlersTestSuite) TestUpdateHandlerIdValidation() {
	id := "trust-me-im-an-id"
	body := TestType{id}

	data, err := json.Marshal(body)
	if err != nil {
		s.FailNow("Unable to marshal request body")
	}
	res := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/test/"+id, bytes.NewReader(data))

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

// The `Update` handler validates the request body.
func (s *HandlersTestSuite) TestUpdateHandlerValidation() {
	uid, _ := uuid.NewRandom()
	id := uid.String()
	body := TestType{"aoeusnth"}

	data, err := json.Marshal(body)
	if err != nil {
		s.FailNow("Unable to marshal request body")
	}
	res := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/test/"+id, bytes.NewReader(data))

	s.router.ServeHTTP(res, req)

	s.Equal(http.StatusBadRequest, res.Result().StatusCode)

	var resBody ErrorResponse
	if err := json.Unmarshal(res.Body.Bytes(), &resBody); err != nil {
		s.FailNow(err.Error())
	}

	expected := ErrorResponse{httputil.Error{
		Code:    "Bad Request",
		Message: "Key: 'TestType.ID' Error:Field validation for 'ID' failed on the 'uuid' tag",
	}}
	s.Equal(expected, resBody)
}

// The `Update` handler returns errors from the controller.
func (s *HandlersTestSuite) TestUpdateHandlerError() {
	uid, _ := uuid.NewRandom()
	id := uid.String()
	body := TestType{id}

	s.controller.On("Update", id).Return(body, httputil.NewError(
		httputil.RecordNotFound,
		errors.New("cannot update what doesn't exist"),
	))

	data, err := json.Marshal(body)
	if err != nil {
		s.FailNow("Unable to marshal request body")
	}
	res := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/test/"+id, bytes.NewReader(data))

	s.router.ServeHTTP(res, req)

	s.Equal(http.StatusNotFound, res.Result().StatusCode)

	var resBody ErrorResponse
	if err := json.Unmarshal(res.Body.Bytes(), &resBody); err != nil {
		s.FailNow(err.Error())
	}

	expected := ErrorResponse{httputil.Error{
		Code:    "Record Not Found",
		Message: "cannot update what doesn't exist",
	}}
	s.Equal(expected, resBody)

	s.controller.AssertExpectations(s.T())
}
