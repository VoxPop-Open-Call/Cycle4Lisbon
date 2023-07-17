package controllers

import (
	"testing"

	"bitbucket.org/pensarmais/cycleforlisbon/src/database/models"
	"bitbucket.org/pensarmais/cycleforlisbon/src/server/access"
	"bitbucket.org/pensarmais/cycleforlisbon/src/util/random"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type ExternalContentControllerTestSuite struct {
	suite.Suite
	external *ExternalContentController
	users    *UserController
	db       *gorm.DB
	acl      *access.ACL
}

// Run each test in a transaction.
func (s *ExternalContentControllerTestSuite) SetupTest() {
	tx := testDb.Begin()
	s.db = tx
	s.external = &ExternalContentController{tx, s.acl}
	s.users = &UserController{tx, s.acl, "", nil}

	err := s.db.Exec("delete from external_contents").Error
	s.Require().NoError(err)
}

// Rollback the transaction after each test.
func (s *ExternalContentControllerTestSuite) TearDownTest() {
	s.db.Rollback()
}

func (s *ExternalContentControllerTestSuite) TestListExternalContents() {
	entries := []models.ExternalContent{
		{
			Title:        "a",
			ArticleUrl:   random.String(100),
			LanguageCode: "en",
		},
		{
			Title:        "b",
			ArticleUrl:   random.String(100),
			State:        "approved",
			LanguageCode: "en",
		},
		{
			Title:        "c",
			ArticleUrl:   random.String(100),
			State:        "pending",
			LanguageCode: "en",
		},
		{
			Title:        "d",
			ArticleUrl:   random.String(100),
			State:        "rejected",
			LanguageCode: "en",
		},
		{
			Title:        "e",
			ArticleUrl:   random.String(100),
			State:        "approved",
			LanguageCode: "en",
		},
	}
	err := s.db.Create(entries).Error
	s.NoError(err)

	// ------------------------------- //
	// User only gets approved entries //
	// ------------------------------- //
	_, ctx, err := createRandomUser(s.users)
	s.NoError(err)

	result, err := s.external.List(ListExternalContentFilters{
		Sort:  Sort{"title asc"},
		State: "rejected", // This should be ignored.
	}, ctx)
	s.NoError(err)
	s.Len(result, 2)
	s.Equal(entries[1].Title, result[0].Title)
	s.Equal(entries[4].Title, result[1].Title)

	// ---------------------------------------- //
	// Admin gets all entries (except rejected) //
	// ---------------------------------------- //
	_, ctx, err = createRandomAdmin(s.users)
	s.NoError(err)

	result, err = s.external.List(ListExternalContentFilters{
		Sort: Sort{"title asc"},
	}, ctx)
	s.NoError(err)
	s.Len(result, 4)
	s.Equal(entries[0].Title, result[0].Title)
	s.Equal(entries[1].Title, result[1].Title)
	s.Equal(entries[2].Title, result[2].Title)
	s.Equal(entries[4].Title, result[3].Title)

	// ----------------------------------- //
	// Admin can query pending events only //
	// ----------------------------------- //
	result, err = s.external.List(ListExternalContentFilters{
		Sort:  Sort{"title asc"},
		State: "pending",
	}, ctx)
	s.NoError(err)
	s.Len(result, 2)
	s.Equal(entries[0].Title, result[0].Title)
	s.Equal(entries[2].Title, result[1].Title)
}

func (s *ExternalContentControllerTestSuite) TestApproveExternalContent() {
	entry := models.ExternalContent{
		Title:        random.String(50),
		ArticleUrl:   random.String(100),
		LanguageCode: "pt",
	}
	err := s.db.Create(&entry).Error
	s.NoError(err)
	s.NotEmpty(entry.ID)

	// -------------------------------------- //
	// Non-admin users cannot approve entries //
	// -------------------------------------- //
	_, ctx, err := createRandomUser(s.users)
	s.NoError(err)
	result, err := s.external.Approve(entry.ID.String(), ctx)
	s.Empty(result)
	s.EqualError(err, "ApiError{"+
		"code: Admin Access Required, "+
		"message: the user must be an administrator to perform this action"+
		"}")

	// -------------------------- //
	// Admins can approve entries //
	// -------------------------- //
	_, ctx, err = createRandomAdmin(s.users)
	s.NoError(err)
	result, err = s.external.Approve(entry.ID.String(), ctx)
	s.NotEmpty(result)
	s.NoError(err)
	s.Equal(entry.ID, result.ID)
	s.Equal(entry.Title, result.Title)
	s.Equal("approved", string(result.State))

	var dbExternalContent models.ExternalContent
	err = s.db.First(&dbExternalContent, "id = ?", entry.ID.String()).Error
	s.NoError(err)
	s.NotEmpty(dbExternalContent)
	s.Equal(entry.ID, result.ID)
	s.Equal(entry.Title, result.Title)
	s.Equal("approved", string(result.State))
}

func (s *ExternalContentControllerTestSuite) TestRejectExternalContent() {
	entry := models.ExternalContent{
		Title:        random.String(50),
		ArticleUrl:   random.String(100),
		LanguageCode: "pt",
	}
	err := s.db.Create(&entry).Error
	s.NoError(err)
	s.NotEmpty(entry.ID)

	// ------------------------------------- //
	// Non-admin users cannot reject entries //
	// ------------------------------------- //
	_, ctx, err := createRandomUser(s.users)
	s.NoError(err)
	result, err := s.external.Reject(entry.ID.String(), ctx)
	s.Empty(result)
	s.EqualError(err, "ApiError{"+
		"code: Admin Access Required, "+
		"message: the user must be an administrator to perform this action"+
		"}")

	// ------------------------- //
	// Admins can reject entries //
	// ------------------------- //
	_, ctx, err = createRandomAdmin(s.users)
	s.NoError(err)
	result, err = s.external.Reject(entry.ID.String(), ctx)
	s.NotEmpty(result)
	s.NoError(err)
	s.Equal(entry.ID, result.ID)
	s.Equal(entry.Title, result.Title)
	s.Equal("rejected", string(result.State))

	var dbExternalContent models.ExternalContent
	err = s.db.First(&dbExternalContent, "id = ?", entry.ID.String()).Error
	s.NoError(err)
	s.NotEmpty(dbExternalContent)
	s.Equal(entry.ID, result.ID)
	s.Equal(entry.Title, result.Title)
	s.Equal("rejected", string(result.State))
}

func TestExternalContentController(t *testing.T) {
	acl := access.New()
	registerAllRules(&ExternalContentController{}, acl)
	registerAllRules(&UserController{}, acl)
	suite.Run(t, &ExternalContentControllerTestSuite{
		acl: acl,
	})
}
