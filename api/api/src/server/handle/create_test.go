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

func (m *MockController) Create(
	params TestType,
	_ *gin.Context,
) (TestType, error) {
	args := m.Called(params)
	return args.Get(0).(TestType), args.Error(1)
}

// The `Create` handler calls the `Create` function of the controller with the
// correct parameters.
func (s *HandlersTestSuite) TestCreateHandler() {
	uid, _ := uuid.NewRandom()
	id := uid.String()
	body := TestType{id}

	s.controller.On("Create", body).Return(body, nil)

	data, err := json.Marshal(body)
	if err != nil {
		s.FailNow("Unable to marshal request body")
	}
	res := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/test", bytes.NewReader(data))

	s.router.ServeHTTP(res, req)

	s.Equal(http.StatusCreated, res.Result().StatusCode)

	var resBody TestType
	if err := json.Unmarshal(res.Body.Bytes(), &resBody); err != nil {
		s.FailNow(err.Error())
	}
	s.Equal(body, resBody)

	s.controller.AssertExpectations(s.T())
}

// The `Create` handler validates the request body.
func (s *HandlersTestSuite) TestCreateHandlerValidation() {
	body := TestType{"invalid-id-123"}

	data, err := json.Marshal(body)
	if err != nil {
		s.FailNow("Unable to marshal request body")
	}
	res := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/test", bytes.NewReader(data))

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

// The `Create` handler retruns errors from the controller.
func (s *HandlersTestSuite) TestCreateHandlerError() {
	uid, _ := uuid.NewRandom()
	id := uid.String()
	body := TestType{id}

	s.controller.On("Create", body).Return(body, errors.New("some error"))

	data, err := json.Marshal(body)
	if err != nil {
		s.FailNow("Unable to marshal request body")
	}
	res := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/test", bytes.NewReader(data))

	s.router.ServeHTTP(res, req)

	s.Equal(http.StatusInternalServerError, res.Result().StatusCode)

	var resBody ErrorResponse
	if err := json.Unmarshal(res.Body.Bytes(), &resBody); err != nil {
		s.FailNow(err.Error())
	}

	expected := ErrorResponse{httputil.Error{
		Code:    "Internal Server Error",
		Message: "some error",
	}}
	s.Equal(expected, resBody)

	s.controller.AssertExpectations(s.T())
}
