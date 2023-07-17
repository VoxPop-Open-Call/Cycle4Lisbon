package controllers

import (
	"testing"

	"bitbucket.org/pensarmais/cycleforlisbon/src/database/models"
	"bitbucket.org/pensarmais/cycleforlisbon/src/server/access"
	"bitbucket.org/pensarmais/cycleforlisbon/src/util/random"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type UserControllerTestSuite struct {
	suite.Suite
	users *UserController
	db    *gorm.DB
	acl   *access.ACL
}

// Run each test in a transaction.
func (s *UserControllerTestSuite) SetupTest() {
	tx := testDb.Begin()
	s.db = tx
	s.users = &UserController{tx, s.acl, "", nil}
}

// Rollback the transaction after each test.
func (s *UserControllerTestSuite) TearDownTest() {
	s.db.Rollback()
}

func (s *UserControllerTestSuite) TestCreateUser() {
	// Create user WITHOUT `Subject`
	params := CreateUserParams{
		Name:  "Hugh Hughes",
		Email: "hugh_hughes@notanemail.org",
	}

	user, err := s.users.Create(params, nil)
	s.NoError(err)
	s.NotEmpty(user)

	s.Equal(params.Name, user.Name)
	s.Equal(params.Email, user.Email)

	s.NotZero(user.ID)
	s.NotZero(user.CreatedAt)
	s.NotZero(user.UpdatedAt)

	s.Equal(user.ID.String(), user.Subject)

	// Create user WITH `Subject`
	params = CreateUserParams{
		Email:   random.String(50),
		Subject: random.String(100),
	}

	user, err = s.users.Create(params, nil)
	s.NoError(err)
	s.NotEmpty(user)
	s.Equal(params.Email, user.Email)
	s.Equal(params.Subject, user.Subject)
}

func (s *UserControllerTestSuite) TestCreateUserDuplicatedEmail() {
	params := CreateUserParams{
		Name:  "some one",
		Email: "some_one@somewhere.com",
	}

	user, err := s.users.Create(params, nil)
	s.NotEmpty(user)
	s.NoError(err)

	user, err = s.users.Create(params, nil)
	s.Empty(user)
	s.EqualError(err,
		"ApiError{"+
			"code: Email Already Registered, "+
			"message: a user with the given email already exists"+
			"}",
	)
}

// Tests the validation of the CreateUserParams.
func (s *UserControllerTestSuite) TestCreateUserParams() {
	testcases := []struct {
		params CreateUserParams
		result string
	}{
		{
			params: CreateUserParams{
				Subject:  random.String(20),
				Password: random.String(20),
				Email:    random.String(20),
			},
			result: "Key: 'CreateUserParams.Subject' " +
				"Error:Field validation for 'Subject' failed on the 'excluded_with' tag\n" +
				"Key: 'CreateUserParams.Password' " +
				"Error:Field validation for 'Password' failed on the 'excluded_with' tag\n" +
				"Key: 'CreateUserParams.Email' Error:Field validation for 'Email' failed on the 'email' tag",
		},
		{
			params: CreateUserParams{},
			result: "Key: 'CreateUserParams.Subject' " +
				"Error:Field validation for 'Subject' failed on the 'required_without' tag\n" +
				"Key: 'CreateUserParams.Password' " +
				"Error:Field validation for 'Password' failed on the 'required_without' tag\n" +
				"Key: 'CreateUserParams.Email' " +
				"Error:Field validation for 'Email' failed on the 'required' tag",
		},
		{
			params: CreateUserParams{
				Password: random.String(20),
				Email:    random.String(10) + "@pensarmais.com",
				Name:     random.String(20),
			},
		},
		{
			params: CreateUserParams{
				Subject: random.String(20),
				Email:   random.String(10) + "@pensarmais.com",
				Name:    random.String(20),
			},
		},
	}

	validate := validator.New()
	validate.SetTagName("binding") // validator uses the `validate` tag by default
	for _, tc := range testcases {
		err := validate.Struct(tc.params)
		if tc.result == "" {
			s.NoError(err)
		} else {
			s.Equal(err.Error(), tc.result)
		}
	}
}

func (s *UserControllerTestSuite) TestProfileGenderValidation() {
	testcases := []struct {
		val string
		res string
	}{
		{val: "M"},
		{val: "F"},
		{val: "X"},
		{val: "m", res: "Key: 'Profile.Gender' " +
			"Error:Field validation for 'Gender' failed on the 'oneof' tag"},
		{val: "f", res: "Key: 'Profile.Gender' " +
			"Error:Field validation for 'Gender' failed on the 'oneof' tag"},
		{val: "1", res: "Key: 'Profile.Gender' " +
			"Error:Field validation for 'Gender' failed on the 'oneof' tag"},
		{val: " ", res: "Key: 'Profile.Gender' " +
			"Error:Field validation for 'Gender' failed on the 'oneof' tag"},
	}

	validate := validator.New()
	validate.SetTagName("binding") // validator uses the `validate` tag by default
	for _, tc := range testcases {
		err := validate.StructPartial(models.Profile{
			Gender: tc.val,
		}, "Gender")
		if tc.res == "" {
			s.NoError(err)
		} else {
			s.Equal(err.Error(), tc.res)

		}
	}

	err := validate.StructPartial(models.Profile{}, "Gender")
	s.NoError(err)
}

func (s *UserControllerTestSuite) TestGetUser() {
	user, ctx, err := createRandomUser(s.users)
	s.NotEmpty(user)
	s.NoError(err)

	result, err := s.users.Get(user.ID.String(), ctx)
	s.NotEmpty(result)
	s.NoError(err)

	// Ignore the `CreatedAt` and `UpdatedAt` fields
	result.CreatedAt = user.CreatedAt
	result.UpdatedAt = user.UpdatedAt

	s.Equal(user, result)
}

func (s *UserControllerTestSuite) TestUpdateUser() {
	user, ctx, err := createRandomUser(s.users)
	s.Require().NoError(err)
	s.Require().NotEmpty(ctx)
	s.Require().NotEmpty(user)

	institution := models.Institution{
		Name:        random.AlphanumericString(20),
		Description: random.AlphanumericString(50),
	}
	err = s.db.Create(&institution).Error
	s.Require().NoError(err)

	initiative := models.Initiative{
		Title:         random.String(50),
		Description:   random.String(200),
		Goal:          uint32(random.Int(200, 1000)),
		EndDate:       "2040-01-01",
		Enabled:       true,
		InstitutionID: institution.ID,
	}
	err = s.db.Create(&initiative).Error
	s.Require().NoError(err)

	params := UpdateUserParams{
		Profile: models.Profile{
			Name:   "Ron Swanson",
			Gender: "M",
		},
		Email:        "boss@parksandrec.gov",
		InitiativeID: &initiative.ID,
	}

	result, err := s.users.Update(user.ID.String(), params, ctx)
	s.NotEmpty(result)
	s.NoError(err)

	result.Initiative.CreatedAt = initiative.CreatedAt
	result.Initiative.UpdatedAt = initiative.UpdatedAt

	s.Equal(user.ID, result.ID)
	s.Equal(params.Email, result.Email)
	s.Equal(params.Profile, result.Profile)
	s.Equal(params.InitiativeID, result.InitiativeID)
	s.Equal(&initiative, result.Initiative)
	s.False(result.Verified)
}

func (s *UserControllerTestSuite) TestListUsers() {
	for i := 0; i < 10; i++ {
		createRandomUser(s.users)
	}

	filters := ListUsersFilters{Pagination: Pagination{
		Limit:  5,
		Offset: 3,
	}}

	users, err := s.users.List(filters, nil)
	s.NoError(err)
	s.Len(users, 5)

	for _, user := range users {
		s.NotEmpty(user)
	}

	filters = ListUsersFilters{
		Pagination: Pagination{10, 0},
		Sort:       Sort{"email asc"},
	}
	users, err = s.users.List(filters, nil)
	s.NoError(err)
	s.Len(users, 10)

	prevEmail := users[0].Email
	for _, u := range users {
		s.GreaterOrEqual(u.Email, prevEmail)
		prevEmail = u.Email
	}
}

func (s *UserControllerTestSuite) TestDeleteUser() {
	user, ctx, err := createRandomUser(s.users)
	s.NotEmpty(user)
	s.NoError(err)

	err = s.users.Delete(user.ID.String(), ctx)
	s.NoError(err)

	_, ctx2, err := createRandomUser(s.users)
	s.NoError(err)

	// Getting a user after deleting should return an error
	result, err := s.users.Get(user.ID.String(), ctx2)
	s.EqualError(err, "ApiError{"+
		"code: Record Not Found, "+
		"message: user not found"+
		"}")
	s.Empty(result)

	// Trying to delete the same user should return an error
	err = s.users.Delete(user.ID.String(), ctx)
	s.EqualError(err, "ApiError{"+
		"code: Token User Not Found, "+
		"message: the token is valid, but the subject doesn't exist"+
		"}")
}

func TestUserController(t *testing.T) {
	acl := access.New()
	registerAllRules(&UserController{}, acl)
	suite.Run(t, &UserControllerTestSuite{
		acl: acl,
	})
}
