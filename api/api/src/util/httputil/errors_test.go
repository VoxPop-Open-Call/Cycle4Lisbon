package httputil

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewApiError(t *testing.T) {
	apiErr := NewError(InternalServerError, errors.New("test"))
	require.NotEmpty(t, apiErr)
	require.Equal(t, 500, apiErr.Status)
	require.Equal(t, "Internal Server Error", apiErr.Code)
	require.Equal(t, "test", apiErr.Message)

	apiErr = NewErrorMsg(BadRequest, "test 1")
	require.NotEmpty(t, apiErr)
	require.Equal(t, 400, apiErr.Status)
	require.Equal(t, "Bad Request", apiErr.Code)
	require.Equal(t, "test 1", apiErr.Message)
}

func TestApiErrorToJson(t *testing.T) {
	apiErr := NewError(RecordNotFound, errors.New("user not found"))
	require.NotEmpty(t, apiErr)

	b, err := json.Marshal(apiErr)
	require.NotEmpty(t, b)
	require.NoError(t, err)

	buff := bytes.Buffer{}
	err = json.Indent(&buff, b, "", "\t")

	exp := "{" + "\n" +
		"\t\"code\": \"Record Not Found\"," + "\n" +
		"\t\"message\": \"user not found\"" + "\n" +
		"}"

	require.Equal(t, exp, buff.String())
}
