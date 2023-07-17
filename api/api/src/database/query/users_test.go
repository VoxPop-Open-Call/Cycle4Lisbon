package query

import (
	"testing"

	"bitbucket.org/pensarmais/cycleforlisbon/src/config"
	"bitbucket.org/pensarmais/cycleforlisbon/src/database"
	"bitbucket.org/pensarmais/cycleforlisbon/src/database/models"
	"bitbucket.org/pensarmais/cycleforlisbon/src/server/middleware"
	"bitbucket.org/pensarmais/cycleforlisbon/src/util/random"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type UserQueriesTestSuite struct {
	suite.Suite
	db *gorm.DB
	tx *gorm.DB
}

// Run each test in a transaction.
func (s *UserQueriesTestSuite) SetupTest() {
	s.tx = s.db.Begin()
}

// Rollback the transaction after each test.
func (s *UserQueriesTestSuite) TearDownTest() {
	s.tx.Rollback()
}

func (s *UserQueriesTestSuite) TestFromClaims() {
	uid, err := uuid.NewRandom()
	s.NoError(err)
	users := []models.User{
		{
			Email:   random.String(30),
			Subject: random.String(30),
		},
		{
			Email:     random.String(30),
			BaseModel: models.BaseModel{ID: uid},
			Subject:   uid.String(),
		},
	}
	err = s.tx.CreateInBatches(users, 5).Error
	s.NoError(err)

	user, err := Users.FromClaims(&middleware.Claims{
		Sub: users[0].Subject,
	}, s.tx)
	s.NoError(err)
	s.NotEmpty(user)
	s.Equal(users[0].Subject, user.Subject)

	user, err = Users.FromClaims(&middleware.Claims{
		Sub:  random.String(30),
		Name: users[1].ID.String(),
	}, s.tx)
	s.NoError(err)
	s.NotEmpty(user)
	s.Equal(users[1].Subject, user.Subject)
	s.Equal(users[1].ID, user.ID)

	user, err = Users.FromClaims(&middleware.Claims{
		Sub:  random.String(50),
		Name: random.String(50),
	}, s.tx)
	s.Empty(user)
	s.EqualError(err, "record not found")
}

func TestUserQueries(t *testing.T) {
	config, err := config.Load("../../../.env")
	require.NoError(t, err)

	db, err := database.Init(config.DbDsn())
	require.NoError(t, err)

	db.Logger = logger.Default.LogMode(logger.Silent)

	suite.Run(t, &UserQueriesTestSuite{db: db})
}
