package controllers

import (
	"net/http/httptest"
	"testing"
	"time"

	"bitbucket.org/pensarmais/cycleforlisbon/src/server/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type FCMControllerTestSuite struct {
	suite.Suite
	fcm   *FCMTokenController
	users *UserController
	db    *gorm.DB
}

// Run each test in a transaction.
func (s *FCMControllerTestSuite) SetupTest() {
	tx := testDb.Begin()
	s.db = tx
	s.fcm = &FCMTokenController{tx}
	s.users = &UserController{tx, nil, "", nil}
}

// Rollback the transaction after each test.
func (s *FCMControllerTestSuite) TearDownTest() {
	s.db.Rollback()
}

func (s *FCMControllerTestSuite) TestCreateToken() {
	user, _, err := createRandomUser(s.users)
	s.NotEmpty(user)
	s.NoError(err)

	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Set(middleware.TokenClaimsKey, middleware.Claims{
		Sub: user.ID.String(),
	})

	params := RegisterFCMTokenParams{"test-fcm-token:123456"}

	token, err := s.fcm.Register(params, ctx)
	s.NotEmpty(token)
	s.NoError(err)

	s.Equal(params.Token, token.Token)
	s.WithinDuration(time.Now(), token.UpdatedAt, time.Second)
	s.WithinDuration(time.Now(), token.CreatedAt, time.Second)

	s.Equal(user.ID, token.UserID)
}

func TestFCMTokenController(t *testing.T) {
	suite.Run(t, &FCMControllerTestSuite{})
}
