package controllers

import (
	"crypto/sha256"
	"math"
	"os"
	"testing"

	"bitbucket.org/pensarmais/cycleforlisbon/src/database/models"
	"bitbucket.org/pensarmais/cycleforlisbon/src/jobs"
	"bitbucket.org/pensarmais/cycleforlisbon/src/server/access"
	"bitbucket.org/pensarmais/cycleforlisbon/src/util/gobutil"
	"bitbucket.org/pensarmais/cycleforlisbon/src/util/latlon"
	"bitbucket.org/pensarmais/cycleforlisbon/src/util/random"
	"bitbucket.org/pensarmais/cycleforlisbon/src/worker"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// Source of the GPX test data: https://github.com/gps-touring/sample-gpx

type TripControllerTestSuite struct {
	suite.Suite
	trips       *TripController
	initiatives *InitiativeController
	users       *UserController
	db          *gorm.DB
	acl         *access.ACL
	wrkr        *MockWorker
	geocoder    *MockGeocoder
	presigner   *MockPresigner
}

type MockWorker struct {
	mock.Mock
}

func (w *MockWorker) Schedule(t *worker.TaskConfig) error {
	return w.Called(t).Error(0)
}

type MockGeocoder struct {
	mock.Mock
}

func (g *MockGeocoder) ReverseAddr(coords latlon.Coords) string {
	return g.Called(coords).String(0)
}

// Run each test in a transaction.
func (s *TripControllerTestSuite) SetupTest() {
	tx := testDb.Begin()
	s.db = tx
	codec := gobutil.NewGobCodec[jobs.UpdateAchievementsArgs]()
	s.trips = &TripController{tx, s.acl, s.wrkr, s.geocoder, codec}
	s.initiatives = &InitiativeController{tx, s.acl, s.presigner}
	s.users = &UserController{tx, s.acl, "", nil}
}

// Rollback the transaction after each test.
func (s *TripControllerTestSuite) TearDownTest() {
	s.db.Rollback()
}

func (s *TripControllerTestSuite) TestListTrips() {
	user1, ctx1, err := createRandomUser(s.users)
	s.Require().NoError(err)
	user2, ctx2, err := createRandomUser(s.users)
	s.Require().NoError(err)

	err = s.db.Exec("delete from trips").Error
	s.Require().NoError(err)

	data0, err := os.ReadFile("./testdata/parcours-morlaix-plougasnou.gpx")
	s.Require().NoError(err)
	hash0 := sha256.New()
	_, err = hash0.Write(data0)
	s.Require().NoError(err)

	data1, err := os.ReadFile("./testdata/1_Roscoff_Morlaix_A_parcours.gpx")
	s.Require().NoError(err)
	hash1 := sha256.New()
	_, err = hash1.Write(data1)
	s.Require().NoError(err)

	trips := []models.Trip{
		{
			GPX:              data0,
			GPXHash:          hash0.Sum(nil),
			Distance:         10.0,
			Duration:         3600,
			DurationInMotion: 3600,
			UserID:           user1.ID,
		},
		{
			GPX:              data1,
			GPXHash:          hash1.Sum(nil),
			Distance:         10.0,
			Duration:         3600,
			DurationInMotion: 3600,
			UserID:           user2.ID,
		},
	}
	err = s.db.Create(&trips).Error
	s.Require().NoError(err)

	res, err := s.trips.List(ListTripsFilters{}, ctx1)
	s.Require().NoError(err)
	s.Len(res, 1)
	res[0].CreatedAt = trips[0].CreatedAt
	res[0].UpdatedAt = trips[0].UpdatedAt
	s.Equal(trips[0].ID, res[0].ID)
	s.Equal(trips[0].GPXHash, res[0].GPXHash)
	s.Equal(trips[0].Distance, res[0].Distance)
	s.Equal(trips[0].Duration, res[0].Duration)

	res, err = s.trips.List(ListTripsFilters{}, ctx2)
	s.Require().NoError(err)
	s.Len(res, 1)
	res[0].CreatedAt = trips[1].CreatedAt
	res[0].UpdatedAt = trips[1].UpdatedAt
	s.Equal(trips[1].ID, res[0].ID)
	s.Equal(trips[1].GPXHash, res[0].GPXHash)
	s.Equal(trips[1].Distance, res[0].Distance)
	s.Equal(trips[1].Duration, res[0].Duration)
}

func (s *TripControllerTestSuite) TestUpload() {
	initiative := models.Initiative{
		Title:       "abc",
		Description: random.String(50),
		Goal:        100_000,
		EndDate:     "2500-01-01", // Should be a while before the test fails.
		Enabled:     true,
		Institution: models.Institution{
			Name:        random.AlphanumericString(20),
			Description: random.AlphanumericString(50),
		},
	}
	err := s.db.Create(&initiative).Error
	s.Require().NoError(err)

	user, ctx, err := createRandomUser(s.users)
	s.Require().NoError(err)
	_, err = s.users.Update(user.ID.String(), UpdateUserParams{
		InitiativeID: &initiative.ID,
	}, ctx)
	s.Require().NoError(err)

	s.wrkr.On("Schedule", mock.AnythingOfType("")).Return(nil)
	s.geocoder.On("ReverseAddr",
		latlon.Coords{Lat: 48.699449859559536, Lon: -3.789234794676304},
	).Return("Starting addr")
	s.geocoder.On("ReverseAddr",
		latlon.Coords{Lat: 48.58171463012695, Lon: -3.831932544708252},
	).Return("Ending addr")

	data, err := os.ReadFile("./testdata/parcours-morlaix-plougasnou.gpx")
	s.Require().NoError(err)
	res, err := s.trips.Upload(data, ctx)
	s.Require().NoError(err)
	s.Equal(initiative.ID, *res.InitiativeID)
	s.Equal(data, res.GPX)
	s.Truef(math.Abs(res.Distance-26.2) < 0.01,
		"distance '%v' not in margin of error", res.Distance)

	s.presigner.On("PresignGetInitiativeImg", initiative.ID.String()).Return(
		"pre-signed url 0", "GET", nil,
	)
	s.presigner.On("PresignGetInstitutionLogo", initiative.Institution.ID.String()).Return(
		"pre-signed url 1", "GET", nil,
	)

	dbInitiative, err := s.initiatives.Get(initiative.ID.String(), ctx)
	s.Require().NoError(err)
	s.Truef(math.Abs(dbInitiative.Credits-26) < 0.01,
		"credits '%v' not in margin of error", dbInitiative.Credits)

	dbUser, err := s.users.Get(user.ID.String(), ctx)
	s.Require().NoError(err)
	s.Equal(uint(1), dbUser.TripCount)
	s.Truef(math.Abs(dbUser.Credits-26) < 0.01,
		"credits '%v' not in margin of error", dbUser.Credits)
	s.Truef(math.Abs(dbUser.TotalDist-26.2) < 0.01,
		"distance '%v' not in margin of error", dbUser.TotalDist)

	s.wrkr.AssertExpectations(s.T())
	s.geocoder.AssertExpectations(s.T())
}

func TestTripController(t *testing.T) {
	acl := access.New()
	registerAllRules(&TripController{}, acl)
	registerAllRules(&InitiativeController{}, acl)
	registerAllRules(&UserController{}, acl)
	suite.Run(t, &TripControllerTestSuite{
		acl:       acl,
		wrkr:      &MockWorker{},
		geocoder:  &MockGeocoder{},
		presigner: &MockPresigner{},
	})
}
