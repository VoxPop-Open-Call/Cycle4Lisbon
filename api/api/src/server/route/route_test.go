package route

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"testing"

	"bitbucket.org/pensarmais/cycleforlisbon/src/aws"
	"bitbucket.org/pensarmais/cycleforlisbon/src/config"
	"bitbucket.org/pensarmais/cycleforlisbon/src/database"
	"bitbucket.org/pensarmais/cycleforlisbon/src/database/models"
	"bitbucket.org/pensarmais/cycleforlisbon/src/server/access"
	"bitbucket.org/pensarmais/cycleforlisbon/src/server/controllers"
	"bitbucket.org/pensarmais/cycleforlisbon/src/server/middleware"
	"bitbucket.org/pensarmais/cycleforlisbon/src/util/httputil"
	"bitbucket.org/pensarmais/cycleforlisbon/src/util/latlon"
	"bitbucket.org/pensarmais/cycleforlisbon/src/util/random"
	"bitbucket.org/pensarmais/cycleforlisbon/src/worker"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/oauth2-proxy/mockoidc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var testDb *gorm.DB

var mockClientIds = []string{"example-app"}

const errMissingAuthHeader = "{\"error\":{" +
	"\"code\":\"Missing Authorization Token\"," +
	"\"message\":\"missing authorization header\"" +
	"}}"

type accessTokenResponse struct {
	Token string `json:"access_token"`
}

// Mocks the login process with `mockoidc`, and adds the token to the
// authorization header of the request.
func addAuthHeader(
	t *testing.T,
	m *mockoidc.MockOIDC,
	request *http.Request,
	user mockoidc.User,
) {
	if user != nil {
		m.QueueUser(user)
	}

	authURL := m.Issuer() + m.AuthorizationEndpoint() + "?" +
		"response_type=code&" +
		"scope=openid%20profile%20email&" +
		"state=somestate&" +
		"redirect_uri=https://app/callback&" +
		"client_id=" + m.ClientID

	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", authURL, nil)
	m.Authorize(res, req)

	// Get the code from the auth response.
	redirect, err := res.Result().Location()
	require.NoError(t, err)
	code := redirect.Query().Get("code")
	require.NotEmpty(t, code)

	tokenURL := m.Issuer() + m.TokenEndpoint() + "?" +
		"grant_type=authorization_code&" +
		"code=" + code + "&" +
		"client_secret=" + m.ClientSecret + "&" +
		"client_id=" + m.ClientID

	res = httptest.NewRecorder()
	req = httptest.NewRequest("GET", tokenURL, nil)
	m.Token(res, req)

	var resBody accessTokenResponse
	if err := json.Unmarshal(res.Body.Bytes(), &resBody); err != nil {
		t.FailNow()
	}

	request.Header.Add("authorization", "bearer "+resBody.Token)
}

func createRandomUser(store *controllers.Store) (models.User, error) {
	params := controllers.CreateUserParams{
		Name:     random.String(15),
		Email:    random.String(7) + "@test.org",
		Password: random.String(10),
	}

	return store.Users.Create(params, nil)
}

// Test that all private endpoints require authentication.
func TestPrivateEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.Default()
	api := router.Group("")
	api.Use(middleware.Error())

	moidc, err := mockoidc.Run()
	require.NoError(t, err, "failed to start mockoidc")
	defer moidc.Shutdown()
	moidc.ClientID = "example-app"
	auth, ch := middleware.Auth(moidc.Config().Issuer, mockClientIds)
	<-ch

	tx := testDb.Begin()
	defer tx.Rollback()

	acl := access.New()
	store := controllers.NewStore(tx,
		acl,
		&worker.Worker{},
		&aws.Client{},
		&latlon.Geocoder{},
		"",
	)

	Users(api, auth, store)
	Password(api, auth, store)
	Initiatives(api, auth, store)
	SDGs(api, auth, store)
	Institutions(api, auth, store)
	Trips(api, auth, store)
	Achievements(api, auth, store)
	POIs(api, auth, store)
	Leaderboard(api, auth, store)
	ExternalContent(api, auth, store)
	FCM(api, auth, store)
	Metrics(api, auth, store)

	user, err := createRandomUser(store)
	require.NotEmpty(t, user)
	require.NoError(t, err)

	uid, err := uuid.NewRandom()
	require.NoError(t, err)

	userData, err := json.Marshal(user)
	require.NoError(t, err)

	testcases := []*http.Request{
		httptest.NewRequest("GET", "/users", nil),
		httptest.NewRequest("GET", "/users/current", nil),
		httptest.NewRequest("GET", "/users/achievements", nil),
		httptest.NewRequest("GET", "/users/"+user.ID.String(), nil),
		httptest.NewRequest("PUT", "/users/"+user.ID.String(), bytes.NewReader(userData)),
		httptest.NewRequest("DELETE", "/users/"+user.ID.String(), nil),

		httptest.NewRequest("GET", "/users/"+user.ID.String()+"/picture-get-url", nil),
		httptest.NewRequest("GET", "/users/"+user.ID.String()+"/picture-put-url", nil),
		httptest.NewRequest("GET", "/users/"+user.ID.String()+"/picture-delete-url", nil),

		httptest.NewRequest("PUT", "/password", nil),

		httptest.NewRequest("GET", "/initiatives", nil),
		httptest.NewRequest("GET", "/initiatives/"+uid.String(), nil),
		httptest.NewRequest("POST", "/initiatives", nil),
		httptest.NewRequest("PUT", "/initiatives/"+uid.String()+"/enable", nil),
		httptest.NewRequest("PUT", "/initiatives/"+uid.String()+"/disable", nil),
		httptest.NewRequest("DELETE", "/initiatives/"+uid.String(), nil),
		httptest.NewRequest("GET", "/initiatives/"+user.ID.String()+"/img-get-url", nil),
		httptest.NewRequest("GET", "/initiatives/"+user.ID.String()+"/img-put-url", nil),
		httptest.NewRequest("GET", "/initiatives/"+user.ID.String()+"/img-delete-url", nil),

		httptest.NewRequest("GET", "/sdgs", nil),

		httptest.NewRequest("GET", "/institutions", nil),
		httptest.NewRequest("GET", "/institutions/"+uid.String(), nil),
		httptest.NewRequest("POST", "/institutions", nil),
		httptest.NewRequest("DELETE", "/institutions/"+uid.String(), nil),
		httptest.NewRequest("GET", "/institutions/"+user.ID.String()+"/logo-get-url", nil),
		httptest.NewRequest("GET", "/institutions/"+user.ID.String()+"/logo-put-url", nil),
		httptest.NewRequest("GET", "/institutions/"+user.ID.String()+"/logo-delete-url", nil),

		httptest.NewRequest("GET", "/trips", nil),
		httptest.NewRequest("GET", "/trips/"+uid.String(), nil),
		httptest.NewRequest("GET", "/trips/"+uid.String()+"/file", nil),
		httptest.NewRequest("POST", "/trips", nil),

		httptest.NewRequest("GET", "/achievements", nil),

		httptest.NewRequest("GET", "/pois", nil),
		httptest.NewRequest("POST", "/pois", nil),

		httptest.NewRequest("GET", "/leaderboard", nil),

		httptest.NewRequest("GET", "/external", nil),
		httptest.NewRequest("PUT", "/external/"+uid.String()+"/approve", nil),
		httptest.NewRequest("PUT", "/external/"+uid.String()+"/reject", nil),

		httptest.NewRequest("POST", "/fcm/register", nil),

		httptest.NewRequest("GET", "/metrics", nil),
	}

	for i, tc := range testcases {
		// No authorization header //
		res := httptest.NewRecorder()
		router.ServeHTTP(res, tc)

		require.Equal(t, http.StatusUnauthorized, res.Result().StatusCode, "failed on testcase %d", i)
		require.Equal(t, errMissingAuthHeader, res.Body.String(), "failed on testcase %d", i)

		// Fake auth token //
		mockToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxIn0.rTCH8cLoGxAm_xw68z-zXVKi9ie6xJn9tnVWjd_9ftE"
		res = httptest.NewRecorder()
		tc.Header.Set("authorization", "bearer "+mockToken)
		router.ServeHTTP(res, tc)

		require.Equal(t, http.StatusUnauthorized, res.Result().StatusCode, "failed on testcase %d", i)
		require.Regexp(
			t,
			regexp.MustCompile("Invalid Authorization Token"),
			res.Body.String(),
			"failed on testcase %d", i,
		)
	}
}

// Test `orderBy` implements custom validation to prevent SQL injection.
func TestOrderByClauseValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.Default()
	api := router.Group("")
	api.Use(middleware.Error())

	moidc, err := mockoidc.Run()
	require.NoError(t, err, "failed to start mockoidc")
	defer moidc.Shutdown()
	moidc.ClientID = "example-app"
	auth, ch := middleware.Auth(moidc.Config().Issuer, mockClientIds)
	<-ch

	tx := testDb.Begin()
	defer tx.Rollback()

	acl := access.New()
	store := controllers.NewStore(
		tx, acl,
		&worker.Worker{},
		&aws.Client{},
		&latlon.Geocoder{},
		"",
	)

	ExternalContent(api, auth, store)

	_, err = createRandomUser(store)
	require.NoError(t, err)

	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/external?orderBy=id%20asc%3Bdrop%20table%20users&type=news", nil)
	addAuthHeader(t, moidc, req, nil)
	router.ServeHTTP(res, req)

	resBody := middleware.ApiError{}
	require.NoError(t, json.Unmarshal(res.Body.Bytes(), &resBody))
	assert.Equal(t, "Bad Request", resBody.Code)
	assert.Equal(t, middleware.ApiError{Error: httputil.Error{
		Code: "Bad Request",
		Message: "Key: 'ListExternalContentFilters.Sort.OrderBy' " +
			"Error:Field validation for 'OrderBy' failed on the 'order_by_clause' tag",
	}}, resBody)
}

// Connect to the database to test the routes.
// Each route test suite should run inside a transaction, to prevent side effects.
func TestMain(m *testing.M) {
	config, err := config.Load("../../../.env")
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	testDb, err = database.Init(config.DbDsn())
	if err != nil {
		log.Fatalf("error connecting to the database: %v", err)
	}

	testDb.Logger = logger.Default.LogMode(logger.Silent)

	os.Exit(m.Run())
}
