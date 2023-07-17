package route

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"bitbucket.org/pensarmais/cycleforlisbon/src/aws"
	"bitbucket.org/pensarmais/cycleforlisbon/src/database/models"
	"bitbucket.org/pensarmais/cycleforlisbon/src/server/access"
	"bitbucket.org/pensarmais/cycleforlisbon/src/server/controllers"
	"bitbucket.org/pensarmais/cycleforlisbon/src/server/middleware"
	"bitbucket.org/pensarmais/cycleforlisbon/src/util/latlon"
	"bitbucket.org/pensarmais/cycleforlisbon/src/util/random"
	"bitbucket.org/pensarmais/cycleforlisbon/src/worker"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/oauth2-proxy/mockoidc"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type FCMRoutesTestSuite struct {
	suite.Suite
	router *gin.Engine
	tx     *gorm.DB
	store  *controllers.Store
	moidc  *mockoidc.MockOIDC
}

func (s *FCMRoutesTestSuite) TestRegisterToken() {
	user, err := createRandomUser(s.store)
	s.NotEmpty(user)
	s.NoError(err)
	mockUser := mockoidc.DefaultUser()
	mockUser.Subject = user.ID.String()

	body := &controllers.RegisterFCMTokenParams{
		Token: "not-a-real-token:1234",
	}

	data, err := json.Marshal(body)
	s.NoError(err)

	res := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/fcm/register", bytes.NewReader(data))
	addAuthHeader(s.T(), s.moidc, req, mockUser)
	s.router.ServeHTTP(res, req)

	s.Equal(http.StatusCreated, res.Result().StatusCode)

	var resBody models.FCMToken
	if err := json.Unmarshal(res.Body.Bytes(), &resBody); err != nil {
		s.FailNow(err.Error())
	}

	s.NotEmpty(resBody)
	s.NotEmpty(resBody.ID)
	s.NotEmpty(resBody.CreatedAt)
	s.NotEmpty(resBody.UpdatedAt)
	s.Equal(user.ID, resBody.UserID)
	s.Equal(body.Token, resBody.Token)
}

func (s *FCMRoutesTestSuite) TestRefreshToken() {
	user, err := createRandomUser(s.store)
	s.NotEmpty(user)
	s.NoError(err)
	uid, err := uuid.Parse(user.ID.String())
	s.NoError(err)

	mockUser := mockoidc.DefaultUser()
	mockUser.Subject = user.ID.String()

	token := &models.FCMToken{
		Token:    random.String(20),
		Failures: 10,
		UserID:   uid,
		BaseModel: models.BaseModel{
			CreatedAt: time.Now().Add(-time.Hour),
			UpdatedAt: time.Now().Add(-time.Hour),
		},
		LastActiveAt: time.Now().Add(-time.Hour),
	}

	result := s.tx.Create(token)
	s.NoError(result.Error)

	body := &controllers.RegisterFCMTokenParams{
		Token: token.Token,
	}

	data, err := json.Marshal(body)
	s.NoError(err)

	res := httptest.NewRecorder()
	req := httptest.NewRequest(
		"POST",
		"/fcm/register",
		bytes.NewReader(data),
	)
	addAuthHeader(s.T(), s.moidc, req, mockUser)
	s.router.ServeHTTP(res, req)

	s.Equal(http.StatusCreated, res.Result().StatusCode)

	var resBody models.FCMToken
	if err := json.Unmarshal(res.Body.Bytes(), &resBody); err != nil {
		s.FailNow(err.Error())
	}

	s.NotEmpty(resBody)
	s.NotEmpty(resBody.ID)
	s.NotEmpty(resBody.CreatedAt)
	s.NotEmpty(resBody.UpdatedAt)
	s.Equal(user.ID, resBody.UserID)
	s.Equal(body.Token, resBody.Token)
	s.WithinDuration(time.Now(), resBody.UpdatedAt, time.Second)

	var dbToken models.FCMToken
	err = s.tx.First(&dbToken, "token = ?", token.Token).Error
	s.NoError(err)
	s.NotEmpty(dbToken)
	s.Equal(resBody.ID, dbToken.ID)
	s.Equal(resBody.Token, dbToken.Token)
	s.Equal(resBody.Failures, dbToken.Failures)
	s.Equal(resBody.UpdatedAt.UTC(), dbToken.UpdatedAt.UTC())
	s.WithinDuration(time.Now().UTC(), dbToken.LastActiveAt.UTC(), time.Second)
}

func TestFCMRoutes(t *testing.T) {
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

	// register fcm routes
	FCM(api, auth, store)

	suite.Run(t, &FCMRoutesTestSuite{
		router: router,
		tx:     tx,
		store:  store,
		moidc:  moidc,
	})

	tx.Rollback()
}
