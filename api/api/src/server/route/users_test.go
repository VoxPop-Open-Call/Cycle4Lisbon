package route

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"bitbucket.org/pensarmais/cycleforlisbon/src/aws"
	"bitbucket.org/pensarmais/cycleforlisbon/src/database/models"
	"bitbucket.org/pensarmais/cycleforlisbon/src/server/access"
	"bitbucket.org/pensarmais/cycleforlisbon/src/server/controllers"
	"bitbucket.org/pensarmais/cycleforlisbon/src/server/middleware"
	"bitbucket.org/pensarmais/cycleforlisbon/src/util/latlon"
	"bitbucket.org/pensarmais/cycleforlisbon/src/worker"
	"github.com/gin-gonic/gin"
	"github.com/oauth2-proxy/mockoidc"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type UserRoutesTestSuite struct {
	suite.Suite
	router *gin.Engine
	store  *controllers.Store
	moidc  *mockoidc.MockOIDC
}

func (s *UserRoutesTestSuite) TestCreateUser() {
	body := &controllers.CreateUserParams{
		Name:     "Eddie Pasana",
		Email:    "ed@test.com",
		Password: "password123",
	}

	data, err := json.Marshal(body)
	s.NoError(err)

	res := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/users", bytes.NewReader(data))
	s.router.ServeHTTP(res, req)

	s.Equal(http.StatusCreated, res.Result().StatusCode)

	var resBody models.User
	if err := json.Unmarshal(res.Body.Bytes(), &resBody); err != nil {
		s.FailNow(err.Error())
	}

	s.NotEmpty(resBody)
	s.NotEmpty(resBody.ID)
	s.NotEmpty(resBody.CreatedAt)
	s.NotEmpty(resBody.UpdatedAt)
	s.Equal("Eddie Pasana", resBody.Name)
	s.Equal("ed@test.com", resBody.Email)
	s.Empty(resBody.HashedPassword)
}

func (s *UserRoutesTestSuite) TestListReturnsAllUsers() {
	_, err := createRandomUser(s.store)
	s.NoError(err)

	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/users", nil)
	addAuthHeader(s.T(), s.moidc, req, nil)
	s.router.ServeHTTP(res, req)

	s.Equal(http.StatusOK, res.Result().StatusCode)

	resBody := []models.User{}
	if err := json.Unmarshal(res.Body.Bytes(), &resBody); err != nil {
		s.FailNow(err.Error())
	}

	for _, u := range resBody {
		s.NotEmpty(u)
		s.NotEmpty(u.ID)
		s.Empty(u.HashedPassword)
	}
}

func TestUserRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.Default()
	api := router.Group("")

	moidc, err := mockoidc.Run()
	require.NoError(t, err, "failed to start mockoidc")
	defer moidc.Shutdown()
	moidc.ClientID = "example-app"
	auth, ch := middleware.Auth(moidc.Config().Issuer, mockClientIds)
	<-ch

	tx := testDb.Begin()
	acl := access.New()
	store := controllers.NewStore(
		tx, acl,
		&worker.Worker{},
		&aws.Client{},
		&latlon.Geocoder{},
		"",
	)

	// register user routes
	Users(api, auth, store)

	suite.Run(t, &UserRoutesTestSuite{
		router: router,
		store:  store,
		moidc:  moidc,
	})

	tx.Rollback()
}
