package middleware

import (
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"bitbucket.org/pensarmais/cycleforlisbon/src/util/httputil"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestHttpUtilError(t *testing.T) {
	res := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(res)
	ctx.Error(httputil.NewErrorMsg(
		httputil.BadRequest,
		"invalid request params",
	))

	Error()(ctx)

	var resBody ApiError
	err := json.Unmarshal(res.Body.Bytes(), &resBody)
	require.NoError(t, err)
	require.NotEmpty(t, resBody)

	require.Equal(t, ApiError{httputil.Error{
		Code:    "Bad Request",
		Message: "invalid request params",
	}}, resBody)
}

func TestInternalError(t *testing.T) {
	res := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(res)
	ctx.Error(errors.New("unexpected error"))

	Error()(ctx)

	var resBody ApiError
	err := json.Unmarshal(res.Body.Bytes(), &resBody)
	require.NoError(t, err)
	require.NotEmpty(t, resBody)

	require.Equal(t, ApiError{httputil.Error{
		Code:    "Internal Server Error",
		Message: "unexpected error",
	}}, resBody)
}

func TestNoError(t *testing.T) {
	res := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(res)

	Error()(ctx)

	require.Empty(t, res.Body.String())
}
