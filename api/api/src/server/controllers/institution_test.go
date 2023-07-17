package controllers

import (
	"testing"

	"bitbucket.org/pensarmais/cycleforlisbon/src/database/models"
	"bitbucket.org/pensarmais/cycleforlisbon/src/server/access"
	"bitbucket.org/pensarmais/cycleforlisbon/src/util/random"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type InstitutionControllerTestSuite struct {
	suite.Suite
	institutions *InstitutionController
	initiatives  *InitiativeController
	users        *UserController
	db           *gorm.DB
	acl          *access.ACL
	presigner    *MockPresigner
}

func (m *MockPresigner) PresignGetInstitutionLogo(institutionID string) (string, string, error) {
	args := m.Called(institutionID)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockPresigner) PresignPutInstitutionLogo(institutionID string) (string, string, error) {
	args := m.Called(institutionID)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockPresigner) PresignDeleteInstitutionLogo(institutionID string) (string, string, error) {
	args := m.Called(institutionID)
	return args.String(0), args.String(1), args.Error(2)
}

// Run each test in a transaction.
func (s *InstitutionControllerTestSuite) SetupTest() {
	tx := testDb.Begin()
	s.db = tx
	s.presigner = &MockPresigner{}
	s.institutions = &InstitutionController{tx, s.acl, s.presigner}
	s.initiatives = &InitiativeController{tx, s.acl, s.presigner}
	s.users = &UserController{tx, s.acl, "", nil}
}

// Rollback the transaction after each test.
func (s *InstitutionControllerTestSuite) TearDownTest() {
	s.presigner.AssertExpectations(s.T())
	s.db.Rollback()
}

func (s *InstitutionControllerTestSuite) TestListInstitutions() {
	err := s.db.Exec("delete from institutions").Error
	s.Require().NoError(err)

	institutions := []models.Institution{
		{
			Name:        "abcdef",
			Description: random.String(500),
		},
		{
			Name:        "ghijkl",
			Description: random.String(500),
		},
	}
	err = s.db.Create(&institutions).Error
	s.Require().NoError(err)

	_, ctx, err := createRandomUser(s.users)
	s.Require().NoError(err)

	s.presigner.On("PresignGetInstitutionLogo", institutions[0].ID.String()).Return(
		"pre-signed url 0", "GET", nil,
	)
	s.presigner.On("PresignGetInstitutionLogo", institutions[1].ID.String()).Return(
		"pre-signed url 1", "GET", nil,
	)

	result, err := s.institutions.List(ListInstitutionsFilters{
		Sort: Sort{"name asc"},
	}, ctx)

	s.NoError(err)
	s.Len(result, 2)
	s.Equal(institutions[0].Name, result[0].Name)
	s.Equal("pre-signed url 0", result[0].PresignedLogoURL)
	s.Equal(institutions[1].Name, result[1].Name)
	s.Equal("pre-signed url 1", result[1].PresignedLogoURL)
}

func (s *InstitutionControllerTestSuite) TestCreateInstitution() {
	params := CreateInstitutionParams{
		Name:        random.AlphanumericString(10),
		Description: random.AlphanumericString(50),
	}

	// ---------------------- //
	// Fails for regular user //
	// ---------------------- //
	_, ctx, err := createRandomUser(s.users)
	s.Require().NoError(err)

	_, err = s.institutions.Create(params, ctx)

	s.EqualError(err, "ApiError{"+
		"code: Admin Access Required, "+
		"message: the user must be an administrator to perform this action"+
		"}")

	// ------------------- //
	// Succeeds for admins //
	// ------------------- //
	_, ctx, err = createRandomAdmin(s.users)
	s.Require().NoError(err)

	result, err := s.institutions.Create(params, ctx)
	s.NoError(err)
	s.Equal(params.Name, result.Name)
	s.Equal(params.Description, result.Description)
}

func (s *InstitutionControllerTestSuite) TestDeleteInstitution() {
	institution := models.Institution{
		Name:        random.AlphanumericString(10),
		Description: random.AlphanumericString(50),
	}
	err := s.db.Create(&institution).Error
	s.Require().NoError(err)

	// ----------------------- //
	// Fails for regular users //
	// ----------------------- //
	_, ctx, err := createRandomUser(s.users)
	s.Require().NoError(err)

	err = s.institutions.Delete(institution.ID.String(), ctx)
	s.EqualError(err, "ApiError{"+
		"code: Admin Access Required, "+
		"message: the user must be an administrator to perform this action"+
		"}")

	// ------------------- //
	// Succeeds for admins //
	// ------------------- //
	_, ctx, err = createRandomAdmin(s.users)
	s.Require().NoError(err)

	err = s.institutions.Delete(institution.ID.String(), ctx)
	s.NoError(err)

	err = s.institutions.Delete(institution.ID.String(), ctx)
	s.EqualError(err, "ApiError{"+
		"code: Record Not Found, "+
		"message: institution not found"+
		"}")
}

func TestInstitutionController(t *testing.T) {
	acl := access.New()
	registerAllRules(&InstitutionController{}, acl)
	registerAllRules(&InitiativeController{}, acl)
	registerAllRules(&UserController{}, acl)
	suite.Run(t, &InstitutionControllerTestSuite{acl: acl})
}
