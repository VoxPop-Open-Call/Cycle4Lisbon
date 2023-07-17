package controllers

import (
	"testing"

	"bitbucket.org/pensarmais/cycleforlisbon/src/database/models"
	"bitbucket.org/pensarmais/cycleforlisbon/src/server/access"
	"bitbucket.org/pensarmais/cycleforlisbon/src/util/random"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type InitiativeControllerTestSuite struct {
	suite.Suite
	initiatives *InitiativeController
	users       *UserController
	db          *gorm.DB
	acl         *access.ACL
	presigner   *MockPresigner
}

func (m *MockPresigner) PresignGetInitiativeImg(initiativeID string) (string, string, error) {
	args := m.Called(initiativeID)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockPresigner) PresignPutInitiativeImg(initiativeID string) (string, string, error) {
	args := m.Called(initiativeID)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockPresigner) PresignDeleteInitiativeImg(initiativeID string) (string, string, error) {
	args := m.Called(initiativeID)
	return args.String(0), args.String(1), args.Error(2)
}

// Run each test in a transaction.
func (s *InitiativeControllerTestSuite) SetupTest() {
	tx := testDb.Begin()
	s.db = tx
	s.presigner = &MockPresigner{}
	s.initiatives = &InitiativeController{tx, s.acl, s.presigner}
	s.users = &UserController{tx, s.acl, "", nil}
}

// Rollback the transaction after each test.
func (s *InitiativeControllerTestSuite) TearDownTest() {
	s.presigner.AssertExpectations(s.T())
	s.db.Rollback()
}

func (s *InitiativeControllerTestSuite) TestListInitiatives() {
	err := s.db.Exec("delete from initiatives").Error
	s.Require().NoError(err)

	initiatives := []models.Initiative{
		{
			Title:       "abc",
			Description: random.String(50),
			Goal:        100_000,
			EndDate:     "2050-01-01",
			Institution: models.Institution{
				Name: random.AlphanumericString(20),
			},
		},
		{
			Title:       "def",
			Description: random.String(50),
			Goal:        100_000,
			EndDate:     "2050-01-01",
			Enabled:     true,
			Institution: models.Institution{
				Name: random.AlphanumericString(20),
			},
		},
	}
	err = s.db.Create(&initiatives).Error
	s.Require().NoError(err)

	// ----------------------------------------------- //
	// Retrieves enabled initiatives for regular users //
	// ----------------------------------------------- //
	_, ctx, err := createRandomUser(s.users)
	s.Require().NoError(err)

	s.presigner.On("PresignGetInitiativeImg", initiatives[0].ID.String()).Return(
		"pre-signed url 0", "GET", nil,
	)
	s.presigner.On("PresignGetInitiativeImg", initiatives[1].ID.String()).Return(
		"pre-signed url 1", "GET", nil,
	)
	s.presigner.On("PresignGetInstitutionLogo", initiatives[0].Institution.ID.String()).Return(
		"pre-signed url 2", "GET", nil,
	)
	s.presigner.On("PresignGetInstitutionLogo", initiatives[1].Institution.ID.String()).Return(
		"pre-signed url 3", "GET", nil,
	)

	result, err := s.initiatives.List(ListInitiativesFilters{
		IncludeDisabled: true, // should be ignored
	}, ctx)

	s.NoError(err)
	s.Len(result, 1)
	s.Equal(initiatives[1].Title, result[0].Title)
	s.Equal("pre-signed url 1", result[0].PresignedImgURL)
	s.Equal("pre-signed url 3", result[0].InstitutionWithImage.PresignedLogoURL)

	// ---------------------------------------- //
	// Includes disabled initiatives for admins //
	// ---------------------------------------- //
	_, ctx, err = createRandomAdmin(s.users)
	s.Require().NoError(err)

	result, err = s.initiatives.List(ListInitiativesFilters{
		IncludeDisabled: true,
		Sort:            Sort{"title asc"},
	}, ctx)

	s.NoError(err)
	s.Len(result, 2)
	s.Equal(initiatives[0].Title, result[0].Title)
	s.Equal("pre-signed url 0", result[0].PresignedImgURL)
	s.Equal(initiatives[1].Title, result[1].Title)
	s.Equal("pre-signed url 1", result[1].PresignedImgURL)
}

func (s *InitiativeControllerTestSuite) TestCreateInitiative() {
	institution := models.Institution{
		Name:        random.String(20),
		Description: random.String(50),
	}
	err := s.db.Create(&institution).Error
	s.Require().NoError(err)

	params := CreateInitiativeParams{
		Title:         random.String(20),
		Description:   random.String(50),
		Goal:          100_000,
		EndDate:       "2050-01-01",
		InstitutionID: institution.ID,
	}

	// ---------------------- //
	// Fails for regular user //
	// ---------------------- //
	_, ctx, err := createRandomUser(s.users)
	s.Require().NoError(err)

	_, err = s.initiatives.Create(params, ctx)

	s.EqualError(err, "ApiError{"+
		"code: Admin Access Required, "+
		"message: the user must be an administrator to perform this action"+
		"}")

	// ------------------- //
	// Succeeds for admins //
	// ------------------- //
	_, ctx, err = createRandomAdmin(s.users)
	s.Require().NoError(err)

	s.presigner.On("PresignGetInitiativeImg", mock.AnythingOfType("string")).Return(
		"pre-signed-url", "GET", nil,
	)
	s.presigner.On("PresignGetInstitutionLogo", institution.ID.String()).Return(
		"pre-signed-url-institution", "GET", nil,
	)

	result, err := s.initiatives.Create(params, ctx)
	s.NoError(err)
	s.Equal(params.Title, result.Title)
	s.Equal(params.Description, result.Description)
	s.Equal(params.Goal, result.Goal)
	s.Equal(params.EndDate, result.EndDate)
	s.Equal(institution.Name, result.Institution.Name)
}

func (s *InitiativeControllerTestSuite) TestEnableInitiative() {
	initiative := models.Initiative{
		Title:       random.String(20),
		Description: random.String(50),
		Goal:        7_000,
		EndDate:     "2050-01-01",
		Institution: models.Institution{
			Name: random.AlphanumericString(20),
		},
	}
	err := s.db.Create(&initiative).Error
	s.Require().NoError(err)

	// ----------------------- //
	// Fails for regular users //
	// ----------------------- //
	_, ctx, err := createRandomUser(s.users)
	s.Require().NoError(err)

	_, err = s.initiatives.Enable(initiative.ID.String(), ctx)
	s.EqualError(err, "ApiError{"+
		"code: Admin Access Required, "+
		"message: the user must be an administrator to perform this action"+
		"}")

	_, err = s.initiatives.Disable(initiative.ID.String(), ctx)
	s.EqualError(err, "ApiError{"+
		"code: Admin Access Required, "+
		"message: the user must be an administrator to perform this action"+
		"}")

	// ------------------- //
	// Succeeds for admins //
	// ------------------- //
	_, ctx, err = createRandomAdmin(s.users)
	s.Require().NoError(err)

	result, err := s.initiatives.Enable(initiative.ID.String(), ctx)
	s.NotEmpty(result)
	s.NoError(err)
	s.True(result.Enabled)

	var dbInitiative models.Initiative
	err = s.db.Find(&dbInitiative, "id = ?", initiative.ID).Error
	s.NoError(err)
	s.True(dbInitiative.Enabled)

	result, err = s.initiatives.Disable(initiative.ID.String(), ctx)
	s.NotEmpty(result)
	s.NoError(err)
	s.False(result.Enabled)

	var dbInitiative2 models.Initiative
	err = s.db.Find(&dbInitiative2, "id = ?", initiative.ID).Error
	s.NoError(err)
	s.False(dbInitiative2.Enabled)
}

func (s *InitiativeControllerTestSuite) TestUpdateInitiative() {
	// TODO
}

func (s *InitiativeControllerTestSuite) TestDeleteInitiative() {
	initiative := models.Initiative{
		Title:       random.String(20),
		Description: random.String(50),
		Goal:        7_000,
		EndDate:     "2050-01-01",
		Institution: models.Institution{
			Name: random.AlphanumericString(20),
		},
	}
	err := s.db.Create(&initiative).Error
	s.Require().NoError(err)

	// ----------------------- //
	// Fails for regular users //
	// ----------------------- //
	_, ctx, err := createRandomUser(s.users)
	s.Require().NoError(err)

	err = s.initiatives.Delete(initiative.ID.String(), ctx)
	s.EqualError(err, "ApiError{"+
		"code: Admin Access Required, "+
		"message: the user must be an administrator to perform this action"+
		"}")

	// ------------------- //
	// Succeeds for admins //
	// ------------------- //
	_, ctx, err = createRandomAdmin(s.users)
	s.Require().NoError(err)

	err = s.initiatives.Delete(initiative.ID.String(), ctx)
	s.NoError(err)

	err = s.initiatives.Delete(initiative.ID.String(), ctx)
	s.EqualError(err, "ApiError{"+
		"code: Record Not Found, "+
		"message: initiative not found"+
		"}")
}

func TestInitiativeController(t *testing.T) {
	acl := access.New()
	registerAllRules(&InitiativeController{}, acl)
	registerAllRules(&UserController{}, acl)
	suite.Run(t, &InitiativeControllerTestSuite{acl: acl})
}
