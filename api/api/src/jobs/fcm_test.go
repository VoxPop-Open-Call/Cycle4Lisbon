package jobs

import (
	"context"
	"errors"
	"testing"
	"time"

	"bitbucket.org/pensarmais/cycleforlisbon/src/config"
	"bitbucket.org/pensarmais/cycleforlisbon/src/database"
	"bitbucket.org/pensarmais/cycleforlisbon/src/database/models"
	"bitbucket.org/pensarmais/cycleforlisbon/src/firebase"
	"bitbucket.org/pensarmais/cycleforlisbon/src/util/gobutil"
	"bitbucket.org/pensarmais/cycleforlisbon/src/util/random"
	"bitbucket.org/pensarmais/cycleforlisbon/src/worker"
	"firebase.google.com/go/v4/messaging"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type mockFcmClient struct {
	mock.Mock
}

func (m *mockFcmClient) Send(
	_ context.Context,
	msg *messaging.Message,
) (string, error) {
	args := m.Called(msg)
	return args.String(0), args.Error(1)
}

func (m *mockFcmClient) SendMulticast(
	_ context.Context,
	msg *messaging.MulticastMessage,
) (*messaging.BatchResponse, error) {
	args := m.Called(msg)
	return args.Get(0).(*messaging.BatchResponse), args.Error(1)
}

func (m *mockFcmClient) SendMulticastDryRun(
	_ context.Context,
	msg *messaging.MulticastMessage,
) (*messaging.BatchResponse, error) {
	args := m.Called(msg)
	return args.Get(0).(*messaging.BatchResponse), args.Error(1)
}

type FCMJobsTestSuite struct {
	suite.Suite
	db         *gorm.DB
	fcm        *mockFcmClient
	wrkr       *worker.Worker
	msgED      *gobutil.GobCodec[messaging.Message]
	mcastMsgED *gobutil.GobCodec[messaging.MulticastMessage]
}

// Send a notification to a topic.
func (s *FCMJobsTestSuite) TestNotifyTopic() {
	job := fcmNotify(s.fcm, s.db)
	s.NotNil(job)
	s.NotEmpty(job)

	msg := messaging.Message{
		Topic: "some-random-topic",
		Data: map[string]string{
			"foo": "bar",
			"baz": "123",
		},
	}
	encodedMsg, err := s.msgED.Encode(msg)
	s.NotEmpty(encodedMsg)
	s.NoError(err)

	s.fcm.On("Send", &msg).Return("all good", nil)
	err = job.Handler(context.Background(), encodedMsg)
	s.NoError(err)

	s.fcm.AssertExpectations(s.T())
}

func createRandomToken(db *gorm.DB, t *testing.T) string {
	user := models.User{
		Email:   random.String(10),
		Subject: random.String(10),
	}
	result := db.Create(&user)
	require.NoError(t, result.Error)
	token := &models.FCMToken{
		UserID:       user.ID,
		Token:        random.String(30),
		LastActiveAt: time.Now().Add(-time.Hour),
	}
	db.Create(token)
	return token.Token
}

// Send a notification to a single token.
func (s *FCMJobsTestSuite) TestNotifyToken() {
	job := fcmNotify(s.fcm, s.db)
	s.NotNil(job)
	s.NotEmpty(job)

	token := createRandomToken(s.db, s.T())
	err := s.db.Model(&models.FCMToken{}).
		Where("token = ?", token).
		Updates(models.FCMToken{
			Failures: 15,
			BaseModel: models.BaseModel{
				UpdatedAt: time.Now().Add(-time.Hour),
			},
		}).Error
	s.NoError(err)

	msg := messaging.Message{
		Token: token,
		Data: map[string]string{
			"foo": "bar",
			"baz": "123",
		},
	}
	encodedMsg, err := s.msgED.Encode(msg)
	s.NotEmpty(encodedMsg)
	s.NoError(err)

	s.fcm.On("Send", &msg).Return("all good", nil)
	err = job.Handler(context.Background(), encodedMsg)
	s.NoError(err)

	var dbToken models.FCMToken
	result := s.db.Model(&models.FCMToken{}).
		First(&dbToken, "token = ?", token)
	s.NoError(result.Error)

	// a successful notification resets the failure counter
	s.Equal(uint16(0), dbToken.Failures)
	s.WithinDuration(time.Now(), dbToken.UpdatedAt, time.Second)

	s.fcm.AssertExpectations(s.T())
}

// Fail to send a notification to a topic.
func (s *FCMJobsTestSuite) TestNotifyTopicError() {
	job := fcmNotify(s.fcm, s.db)
	s.NotNil(job)
	s.NotEmpty(job)

	msg := messaging.Message{
		Topic: "another-topic",
		Data: map[string]string{
			"foo": "bar",
			"baz": "123",
		},
	}
	encodedMsg, err := s.msgED.Encode(msg)
	s.NotEmpty(encodedMsg)
	s.NoError(err)

	s.fcm.On("Send", &msg).Return("", errors.New("invalid topic"))
	err = job.Handler(context.Background(), encodedMsg)
	s.Error(err)
	s.Equal("error sending message: invalid topic", err.Error())

	s.fcm.AssertExpectations(s.T())
}

// Fail to send a notification to a single token.
func (s *FCMJobsTestSuite) TestNotifyIncrementsFailureCount() {
	job := fcmNotify(s.fcm, s.db)
	s.NotNil(job)
	s.NotEmpty(job)

	token := createRandomToken(s.db, s.T())

	msg := messaging.Message{
		Token: token,
		Data: map[string]string{
			"foo": "bar",
			"baz": "123",
		},
	}
	encodedMsg, err := s.msgED.Encode(msg)
	s.NotEmpty(encodedMsg)
	s.NoError(err)

	s.fcm.On("Send", &msg).Return("ups", errors.New("nope"))
	err = job.Handler(context.Background(), encodedMsg)
	s.Equal("error sending message: nope", err.Error())

	var dbToken models.FCMToken
	result := s.db.Model(&models.FCMToken{}).
		First(&dbToken, "token = ?", token)
	s.NoError(result.Error)

	s.Equal(uint16(1), dbToken.Failures)

	s.fcm.AssertExpectations(s.T())
}

// Send a multicast notification.
func (s *FCMJobsTestSuite) TestMulticast() {
	tokens := []string{
		createRandomToken(s.db, s.T()),
		createRandomToken(s.db, s.T()),
		createRandomToken(s.db, s.T()),
	}

	job := fcmMulticast(s.fcm, s.db)
	s.NotNil(job)
	s.NotEmpty(job)

	msg := messaging.MulticastMessage{
		Tokens: tokens,
		Data: map[string]string{
			"foo": "bar",
			"baz": "123",
		},
	}
	encodedMsg, err := s.mcastMsgED.Encode(msg)
	s.NotEmpty(encodedMsg)
	s.NoError(err)

	s.fcm.On("SendMulticast", &msg).Return(&messaging.BatchResponse{
		SuccessCount: 3,
		FailureCount: 0,
		Responses: []*messaging.SendResponse{
			{Success: true},
			{Success: true},
			{Success: true},
		},
	}, nil)
	err = job.Handler(context.Background(), encodedMsg)
	s.NoError(err)

	for _, token := range tokens {
		var dbToken models.FCMToken
		result := s.db.Model(&models.FCMToken{}).
			First(&dbToken, "token = ?", token)
		s.NoError(result.Error)

		s.Equal(uint16(0), dbToken.Failures)
		s.WithinDuration(time.Now(), dbToken.UpdatedAt, time.Second)
		s.WithinDuration(time.Now(), dbToken.LastActiveAt, time.Second)
	}
}

// Send a partially successful multicast message (some tokens fail).
func (s *FCMJobsTestSuite) TestMulticastPartialSuccess() {
	tokens := []string{
		createRandomToken(s.db, s.T()),
		createRandomToken(s.db, s.T()),
		createRandomToken(s.db, s.T()),
	}

	job := fcmMulticast(s.fcm, s.db)
	s.NotNil(job)
	s.NotEmpty(job)

	msg := messaging.MulticastMessage{
		Tokens: tokens,
		Data: map[string]string{
			"foo": "bar",
			"baz": "123",
		},
	}
	encodedMsg, err := s.mcastMsgED.Encode(msg)
	s.NotEmpty(encodedMsg)
	s.NoError(err)

	s.fcm.On("SendMulticast", &msg).Return(&messaging.BatchResponse{
		SuccessCount: 2,
		FailureCount: 1,
		Responses: []*messaging.SendResponse{
			{Success: false},
			{Success: true},
			{Success: true},
		},
	}, nil)
	err = job.Handler(context.Background(), encodedMsg)
	s.NoError(err)

	var dbToken models.FCMToken
	result := s.db.Model(&models.FCMToken{}).
		First(&dbToken, "token = ?", tokens[0])
	s.NoError(result.Error)

	s.Equal(uint16(1), dbToken.Failures)

	for i := 1; i < 3; i++ {
		var dbToken models.FCMToken
		result := s.db.Model(&models.FCMToken{}).
			First(&dbToken, "token = ?", tokens[i])
		s.NoError(result.Error)

		s.Equal(uint16(0), dbToken.Failures)
		s.WithinDuration(time.Now(), dbToken.UpdatedAt, time.Second)
	}
}

// The cleanup job tests and deletes invalid tokens.
func (s *FCMJobsTestSuite) TestCleanup() {
	job := fcmCleanup(s.wrkr, s.fcm, s.db)
	s.NotNil(job)
	s.NotEmpty(job)

	user := models.User{
		Email:   random.String(10),
		Subject: random.String(30),
	}
	err := s.db.Create(&user).Error
	s.NoError(err)

	err = s.db.Exec("delete from fcm_tokens").Error
	s.NoError(err)

	tokens := []models.FCMToken{
		{
			UserID:   user.ID,
			Token:    random.String(20),
			Failures: 100,
		},
		{
			UserID:       user.ID,
			Token:        random.String(20),
			LastActiveAt: time.Now().Add(-100 * 24 * time.Hour),
		},
		{
			UserID:       user.ID,
			Token:        random.String(20),
			LastActiveAt: time.Now().Add(-61 * 24 * time.Hour),
		},
		{
			UserID:       user.ID,
			Token:        random.String(20),
			Failures:     150,
			LastActiveAt: time.Now().Add(-50 * 24 * time.Hour),
		},
		{
			UserID: user.ID,
			Token:  random.String(20),
		},
	}

	s.db.CreateInBatches(tokens, 5)

	expMsg := firebase.NewTestMessage([]string{
		tokens[0].Token,
		tokens[3].Token,
	})

	s.fcm.On("SendMulticastDryRun", &expMsg).Return(&messaging.BatchResponse{
		SuccessCount: 1,
		FailureCount: 1,
		Responses: []*messaging.SendResponse{
			{Success: true},
			{Success: false},
		},
	}, nil)

	err = job.Handler(context.Background(), nil)
	s.NoError(err)

	var dbTokens []models.FCMToken
	err = s.db.Model(&dbTokens).Find(&dbTokens).Error

	s.Len(dbTokens, 2)
	s.Equal(uint16(0), dbTokens[0].Failures)
	s.Equal(uint16(0), dbTokens[1].Failures)

	var token models.FCMToken
	err = s.db.Where("token = ?", tokens[1].Token).First(&token).Error
	s.Equal(gorm.ErrRecordNotFound, err)
	s.Empty(token)
	err = s.db.Where("token = ?", tokens[2].Token).First(&token).Error
	s.Equal(gorm.ErrRecordNotFound, err)
	s.Empty(token)
	err = s.db.Where("token = ?", tokens[3].Token).First(&token).Error
	s.Equal(gorm.ErrRecordNotFound, err)
	s.Empty(token)

	s.fcm.AssertExpectations(s.T())
}

func TestFCMJobs(t *testing.T) {
	config, err := config.Load("../../.env")
	require.NoError(t, err)

	db, err := database.Init(config.DbDsn())
	require.NoError(t, err)
	db.Logger = logger.Default.LogMode(logger.Silent)

	tx := db.Begin()
	wrkr := worker.New(worker.NewDbQueue(tx))
	msgED := gobutil.NewGobCodec[messaging.Message]()
	mcastMsgED := gobutil.NewGobCodec[messaging.MulticastMessage]()

	suite.Run(t, &FCMJobsTestSuite{
		db:         tx,
		fcm:        &mockFcmClient{},
		wrkr:       wrkr,
		msgED:      msgED,
		mcastMsgED: mcastMsgED,
	})

	tx.Rollback()
}
