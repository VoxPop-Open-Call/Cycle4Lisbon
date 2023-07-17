package controllers

import (
	"net/http"
	"testing"
	"time"

	"bitbucket.org/pensarmais/cycleforlisbon/src/database/models"
	"bitbucket.org/pensarmais/cycleforlisbon/src/util/password"
	"bitbucket.org/pensarmais/cycleforlisbon/src/util/random"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type MockEmailer struct {
	mock.Mock
}

func (e *MockEmailer) SendPasswordResetEmail(email, code string) error {
	return e.Called(email, code).Error(0)
}

func (e *MockEmailer) SendPasswordChangedEmail(email string) error {
	return e.Called(email).Error(0)
}

type PasswordControllerTestSuite struct {
	suite.Suite
	passwd  *PasswordController
	users   *UserController
	emailer *MockEmailer
	db      *gorm.DB
}

// Run each test in a transaction.
func (s *PasswordControllerTestSuite) SetupTest() {
	tx := testDb.Begin()
	s.db = tx
	s.emailer = &MockEmailer{}
	s.passwd = &PasswordController{tx, s.emailer}
	s.users = &UserController{tx, nil, "", nil}
}

// Rollback the transaction after each test.
func (s *PasswordControllerTestSuite) TearDownTest() {
	s.db.Rollback()
}

func (s *PasswordControllerTestSuite) TestRequestReset0() {
	// -------------------------------------------- //
	// Fails silently if the email isn't registered //
	// -------------------------------------------- //
	status, err := s.passwd.RequestReset(RequestPasswordResetParams{
		Email: "definitelynotanemail@hello.io",
	}, nil)

	s.Equal(http.StatusAccepted, status)
	s.NoError(err)

	s.emailer.AssertExpectations(s.T())
}

func (s *PasswordControllerTestSuite) TestRequestReset1() {
	// ---------------------------------------- //
	// Creates the record and calls the emailer //
	// ---------------------------------------- //
	user, _, err := createRandomUser(s.users)
	s.Require().NoError(err)

	var code string
	s.emailer.
		On("SendPasswordResetEmail", user.Email, mock.MatchedBy(func(c string) bool {
			// I think this is the only way to match any code (because the
			// value is random) and still be able to retrieve it for testing.
			code = c
			return true
		})).
		Return(nil)

	status, err := s.passwd.RequestReset(RequestPasswordResetParams{
		Email: user.Email,
	}, nil)

	s.Equal(http.StatusAccepted, status)
	s.NoError(err)

	var record models.PasswordResetCode
	err = s.db.First(&record, "email = ?", user.Email).Error
	s.Require().NoError(err)

	s.False(record.Used)
	s.Equal(code, record.Code)
	s.Equal(user.Email, record.Email, code)
	s.WithinDuration(time.Now(), record.CreatedAt, time.Second)
	s.WithinDuration(time.Now().Add(resetCodeLifetime), record.ExpiresAt, time.Second)

	s.emailer.AssertExpectations(s.T())
}

func (s *PasswordControllerTestSuite) TestConfirmReset0() {
	// ------------------------ //
	// Reset code doesn't exist //
	// ------------------------ //
	_, err := s.passwd.ConfirmReset(ConfirmPasswordResetParams{
		Code: "12345678912345678912345678912345",
		New:  "pass@word",
	}, nil)

	s.EqualError(err, "ApiError{code: Record Not Found, message: reset code not found}")

	s.emailer.AssertExpectations(s.T())
}

func (s *PasswordControllerTestSuite) TestConfirmReset1() {
	// --------------------- //
	// Reset code is expired //
	// --------------------- //
	record := &models.PasswordResetCode{
		Code:      random.String(32),
		Email:     "abcdef@ghijk.lmnop",
		CreatedAt: time.Now().Add(-10 * time.Hour),
		ExpiresAt: time.Now().Add(-9 * time.Hour),
		Used:      false,
	}
	err := s.db.Create(record).Error
	s.Require().NoError(err)
	var records []models.PasswordResetCode
	s.db.Find(&records)

	_, err = s.passwd.ConfirmReset(ConfirmPasswordResetParams{
		Code: record.Code,
		New:  "pass@word",
	}, nil)

	s.EqualError(err, "ApiError{"+
		"code: Password Reset Code Expired, "+
		"message: the password reset code is no longer valid"+
		"}")

	s.emailer.AssertExpectations(s.T())
}

func (s *PasswordControllerTestSuite) TestConfirmReset2() {
	// ------------------ //
	// Reset code is used //
	// ------------------ //
	record := &models.PasswordResetCode{
		Code:      random.String(32),
		Email:     "abcdef@ghijk.lmnop",
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(1 * time.Hour),
		Used:      true,
	}
	err := s.db.Create(record).Error
	s.Require().NoError(err)
	var records []models.PasswordResetCode
	s.db.Find(&records)

	_, err = s.passwd.ConfirmReset(ConfirmPasswordResetParams{
		Code: record.Code,
		New:  "pass@word",
	}, nil)

	s.EqualError(err, "ApiError{"+
		"code: Password Reset Code Already Used, "+
		"message: the password reset code has already been used"+
		"}")

	s.emailer.AssertExpectations(s.T())
}

func (s *PasswordControllerTestSuite) TestConfirmReset3() {
	// --------------------------------------------------------- //
	// User doesn't exist                                        //
	// (shouldn't happen in practice, just to test the rollback) //
	// --------------------------------------------------------- //
	record := &models.PasswordResetCode{
		Code:      random.String(32),
		Email:     "abcdef@ghijk.lmnop",
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(1 * time.Hour),
		Used:      false,
	}
	err := s.db.Create(record).Error
	s.Require().NoError(err)
	var records []models.PasswordResetCode
	s.db.Find(&records)

	_, err = s.passwd.ConfirmReset(ConfirmPasswordResetParams{
		Code: record.Code,
		New:  "pass@word",
	}, nil)

	s.EqualError(err, "ApiError{code: Record Not Found, message: user not found}")

	s.emailer.AssertExpectations(s.T())
}

func (s *PasswordControllerTestSuite) TestConfirmReset4() {
	// ----------------------------- //
	// All good, password is updated //
	// ----------------------------- //
	user, _, err := createRandomUser(s.users)
	s.Require().NoError(err)

	record := &models.PasswordResetCode{
		Code:      random.String(32),
		Email:     user.Email,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(1 * time.Hour),
		Used:      false,
	}
	err = s.db.Create(record).Error
	s.Require().NoError(err)
	var records []models.PasswordResetCode
	s.db.Find(&records)

	s.emailer.On("SendPasswordChangedEmail", user.Email).Return(nil)

	status, err := s.passwd.ConfirmReset(ConfirmPasswordResetParams{
		Code: record.Code,
		New:  "pass@word",
	}, nil)

	s.Equal(http.StatusNoContent, status)
	s.NoError(err)

	var dbRecord models.PasswordResetCode
	err = s.db.First(&dbRecord, "email = ?", user.Email).Error
	s.Require().NoError(err)
	s.True(dbRecord.Used)

	var dbUser models.User
	err = s.db.First(&dbUser, "id = ?", user.ID).Error
	s.Require().NoError(err)
	s.True(password.Check("pass@word", dbUser.HashedPassword))

	s.emailer.AssertExpectations(s.T())
}

func (s *PasswordControllerTestSuite) TestUpdatePassword() {
	_, ctx, err := createRandomUser(s.users)
	s.Require().NoError(err)

	status, err := s.passwd.Update(UpdatePasswordParams{
		Old: "pass@word",
		New: random.AlphanumericString(20),
	}, ctx)
	s.Equal(http.StatusNoContent, status)
	s.Require().NoError(err)

	status, err = s.passwd.Update(UpdatePasswordParams{
		Old: "pass@word",
		New: random.AlphanumericString(20),
	}, ctx)
	s.EqualError(err, "ApiError{"+
		"code: Incorrect Password, "+
		"message: the user's current password doesn't match the one provided"+
		"}")
}

func TestPasswordController(t *testing.T) {
	suite.Run(t, &PasswordControllerTestSuite{})
}
