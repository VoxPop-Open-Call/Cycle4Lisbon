package middleware

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/oauth2-proxy/mockoidc"
	"github.com/stretchr/testify/require"
)

const mockToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxIn0.rTCH8cLoGxAm_xw68z-zXVKi9ie6xJn9tnVWjd_9ftE"

var mockClientIds = []string{"example-app"}

func TestGetBearerToken(t *testing.T) {
	var testcases = []struct {
		header string
		result string
		err    error
	}{
		{"", "", errNoAuthorizationHeader},
		{"Basic dGVzdDoxMjPCow==", "", errors.New("auth type 'Basic' not supported")},
		{"Bearer " + mockToken, mockToken, nil},
	}

	gin.SetMode(gin.TestMode)
	for _, tc := range testcases {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest("POST", "/", nil)
		c.Request.Header.Add("authorization", tc.header)

		res, err := getBearerToken(c)

		require.Equal(t, tc.result, res)
		require.Equal(t, tc.err, err)
	}
}

func TestKnownClientId(t *testing.T) {
	testcases := []struct {
		id  string
		ids []string
		res bool
	}{
		{
			id:  "app",
			ids: []string{"someapp", "app", "another-app"},
			res: true,
		},
		{
			id:  "test-app",
			ids: []string{"someapp", "app", "another-app"},
			res: false,
		},
		{
			id:  "",
			ids: []string{"someapp", "app", "another-app"},
			res: false,
		},
		{
			id:  "app",
			ids: []string{},
			res: false,
		},
	}

	for _, tc := range testcases {
		require.Equal(t, tc.res, knownClientId(tc.id, tc.ids))
	}
}

func TestAuthHandler(t *testing.T) {
	moidc, err := mockoidc.Run()
	require.NoError(t, err, "failed to start mockoidc")
	defer moidc.Shutdown()
	moidc.ClientID = "example-app"

	auth, ch := Auth(moidc.Config().Issuer, mockClientIds)
	<-ch

	res := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(res)
	req, err := http.NewRequest("GET", "/", nil)
	require.NoError(t, err)
	c.Request = req
	addAuthHeader(t, moidc, c.Request)

	auth(c)

	val, ok := c.Get(TokenClaimsKey)
	require.True(t, ok)
	require.NotEmpty(t, val)

	claims := val.(Claims)
	require.Equal(t, moidc.Issuer(), claims.Issuer)
	require.Equal(t, moidc.ClientID, claims.ClientID)
	require.Equal(t, mockoidc.DefaultUser().ID(), claims.Sub)
}

func TestFakeToken(t *testing.T) {
	moidc, err := mockoidc.Run()
	require.NoError(t, err, "failed to start mockoidc")
	defer moidc.Shutdown()
	moidc.ClientID = "example-app"
	moidc.AccessTTL = time.Minute
	auth, ch := Auth(moidc.Config().Issuer, mockClientIds)
	<-ch

	res := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(res)
	req, err := http.NewRequest("GET", "/", nil)
	require.NoError(t, err)
	req.Header.Add("authorization", "bearer "+mockToken)
	c.Request = req

	auth(c)

	require.Equal(t, http.StatusUnauthorized, res.Result().StatusCode)
	require.Regexp(
		t,
		regexp.MustCompile("oidc: id token issued by a different provider"),
		res.Body.String(),
	)

	val, ok := c.Get(TokenClaimsKey)
	require.Nil(t, val)
	require.False(t, ok)
}

func TestUnknownClient(t *testing.T) {
	moidc, err := mockoidc.Run()
	require.NoError(t, err, "failed to start mockoidc")
	defer moidc.Shutdown()
	moidc.ClientID = "unknown-client-id"

	auth, ch := Auth(moidc.Config().Issuer, mockClientIds)
	<-ch

	res := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(res)
	req, err := http.NewRequest("GET", "/", nil)
	require.NoError(t, err)
	c.Request = req
	addAuthHeader(t, moidc, c.Request)

	auth(c)

	require.Equal(t, http.StatusUnauthorized, res.Result().StatusCode)
	require.Equal(
		t,
		"{\"error\":{"+
			"\"code\":\"Invalid Authorization Token\","+
			"\"message\":\"unkown client id\""+
			"}}",
		res.Body.String(),
	)

	val, ok := c.Get(TokenClaimsKey)
	require.Nil(t, val)
	require.False(t, ok)
}

type accessTokenResponse struct {
	Token string `json:"access_token"`
}

// Mocks the login process with `mockoidc`, and adds the token to the
// authorization header of the request.
func addAuthHeader(t *testing.T, m *mockoidc.MockOIDC, request *http.Request) {
	authUrl := m.Issuer() + m.AuthorizationEndpoint() + "?" +
		"response_type=code&" +
		"scope=openid%20profile%20email&" +
		"state=somestate&" +
		"redirect_uri=https://app/callback&" +
		"client_id=" + m.ClientID

	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", authUrl, nil)
	m.Authorize(res, req)

	// Get the code from the auth response
	redirect, err := res.Result().Location()
	require.NoError(t, err)
	code := redirect.Query().Get("code")
	require.NotEmpty(t, code)

	tokenUrl := m.Issuer() + m.TokenEndpoint() + "?" +
		"grant_type=authorization_code&" +
		"code=" + code + "&" +
		"client_secret=" + m.ClientSecret + "&" +
		"client_id=" + m.ClientID

	res = httptest.NewRecorder()
	req = httptest.NewRequest("GET", tokenUrl, nil)
	m.Token(res, req)

	var resBody accessTokenResponse
	if err := json.Unmarshal(res.Body.Bytes(), &resBody); err != nil {
		t.FailNow()
	}

	request.Header.Add("authorization", "bearer "+resBody.Token)
}
