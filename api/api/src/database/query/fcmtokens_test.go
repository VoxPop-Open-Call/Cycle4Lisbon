package query

import (
	"testing"
	"time"

	"bitbucket.org/pensarmais/cycleforlisbon/src/config"
	"bitbucket.org/pensarmais/cycleforlisbon/src/database"
	"bitbucket.org/pensarmais/cycleforlisbon/src/database/models"
	"bitbucket.org/pensarmais/cycleforlisbon/src/util/random"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type FCMTokenQueriesTestSuite struct {
	suite.Suite
	db   *gorm.DB
	user *models.User
}

func (s *FCMTokenQueriesTestSuite) TestIncrementFailureCount() {
	token := &models.FCMToken{
		Token:        random.String(30),
		Failures:     25,
		LastActiveAt: time.Now().Add(-24 * time.Hour),
		UserID:       s.user.ID,
	}
	err := s.db.Create(token).Error
	s.NoError(err)

	err = FCMTokens.IncrementFailureCount(token.Token, s.db)
	s.NoError(err)

	var dbToken models.FCMToken
	err = s.db.First(&dbToken, "id = ?", token.ID.String()).Error
	s.NoError(err)
	s.NotEmpty(dbToken)
	s.Equal(token.ID, dbToken.ID)
	s.Equal(token.Token, dbToken.Token)
	s.Equal(uint16(26), dbToken.Failures)
	s.WithinDuration(token.LastActiveAt, dbToken.LastActiveAt, time.Second)
	s.WithinDuration(time.Now(), dbToken.UpdatedAt, time.Second)
	s.Equal(s.user.ID, dbToken.UserID)
}

func (s *FCMTokenQueriesTestSuite) TestResetFailureCount() {
	token := &models.FCMToken{
		Token:        random.String(30),
		Failures:     25,
		LastActiveAt: time.Now().Add(-24 * time.Hour),
		UserID:       s.user.ID,
	}
	err := s.db.Create(token).Error
	s.NoError(err)

	err = FCMTokens.ResetFailureCount(token.Token, s.db)
	s.NoError(err)

	var dbToken models.FCMToken
	err = s.db.First(&dbToken, "id = ?", token.ID.String()).Error
	s.NoError(err)
	s.NotEmpty(dbToken)
	s.Equal(token.ID, dbToken.ID)
	s.Equal(token.Token, dbToken.Token)
	s.Equal(uint16(0), dbToken.Failures)
	s.WithinDuration(time.Now(), dbToken.LastActiveAt, time.Second)
	s.WithinDuration(time.Now(), dbToken.UpdatedAt, time.Second)
	s.Equal(s.user.ID, dbToken.UserID)
}

func (s *FCMTokenQueriesTestSuite) TestDelete() {
	token := &models.FCMToken{
		Token:        random.String(30),
		Failures:     25,
		LastActiveAt: time.Now().Add(-24 * time.Hour),
		UserID:       s.user.ID,
	}
	err := s.db.Create(token).Error
	s.NoError(err)

	err = FCMTokens.Delete(token.Token, s.db)
	s.NoError(err)

	var dbToken models.FCMToken
	err = s.db.First(&dbToken, "id = ?", token.ID.String()).Error
	s.Error(err)
	s.Equal(err, gorm.ErrRecordNotFound)
}

func (s *FCMTokenQueriesTestSuite) TestDeleteInactive() {
	err := s.db.Exec("delete from fcm_tokens").Error
	s.NoError(err)

	tokens := []models.FCMToken{
		{
			Token:        random.String(30),
			Failures:     25,
			LastActiveAt: time.Now().Add(-24 * time.Hour),
			UserID:       s.user.ID,
		},
		{
			Token:        random.String(30),
			Failures:     25,
			LastActiveAt: time.Now().Add(-120 * 24 * time.Hour),
			UserID:       s.user.ID,
		},
	}
	err = s.db.CreateInBatches(tokens, 5).Error
	s.NoError(err)

	count, err := FCMTokens.DeleteAllInactive(60*24*time.Hour, s.db)
	s.NoError(err)

	s.Equal(1, count)
	var dbTokens []models.FCMToken
	err = s.db.Find(&dbTokens).Error
	s.NoError(err)
	s.Len(dbTokens, 1)
	s.Equal(tokens[0].Token, dbTokens[0].Token)
}

func (s *FCMTokenQueriesTestSuite) TestGetFailed() {
	err := s.db.Exec("delete from fcm_tokens").Error
	s.NoError(err)

	tokens := []models.FCMToken{
		{
			Token:    random.String(30),
			Failures: 25,
			UserID:   s.user.ID,
		},
		{
			Token:    random.String(30),
			Failures: 16,
			UserID:   s.user.ID,
		},
		{
			Token:    random.String(30),
			Failures: 42,
			UserID:   s.user.ID,
		},
	}
	err = s.db.CreateInBatches(tokens, 5).Error
	s.NoError(err)

	failed, err := FCMTokens.GetFailed(20, s.db)
	s.NoError(err)
	s.Len(failed, 2)
	s.Equal(tokens[0].Token, failed[0].Token)
	s.Equal(tokens[2].Token, failed[1].Token)
}

func TestFCMTokenQueries(t *testing.T) {
	config, err := config.Load("../../../.env")
	require.NoError(t, err)

	db, err := database.Init(config.DbDsn())
	require.NoError(t, err)

	db.Logger = logger.Default.LogMode(logger.Silent)

	tx := db.Begin()

	testUser := &models.User{}
	tx.Create(testUser)

	suite.Run(t, &FCMTokenQueriesTestSuite{
		db:   tx,
		user: testUser,
	})

	tx.Rollback()
}
