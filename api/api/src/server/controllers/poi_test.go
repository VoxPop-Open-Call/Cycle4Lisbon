package controllers

import (
	"net/http/httptest"
	"os"
	"testing"

	"bitbucket.org/pensarmais/cycleforlisbon/src/database/models"
	"bitbucket.org/pensarmais/cycleforlisbon/src/server/access"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// Source of the GPX test data: https://github.com/gps-touring/sample-gpx

type POIControllerTestSuite struct {
	suite.Suite
	users *UserController
	pois  *POIController
	db    *gorm.DB
	acl   *access.ACL
}

// Run each test in a transaction.
func (s *POIControllerTestSuite) SetupTest() {
	tx := testDb.Begin()
	s.db = tx
	s.users = &UserController{tx, s.acl, "", nil}
	s.pois = &POIController{tx, s.acl}
}

// Rollback the transaction after each test.
func (s *POIControllerTestSuite) TearDownTest() {
	s.db.Rollback()
}

func (s *POIControllerTestSuite) TestImport() {
	err := s.db.Exec("delete from points_of_interest").Error
	s.Require().NoError(err)

	_, ctx, err := createRandomAdmin(s.users)
	s.Require().NoError(err)

	data, err := os.ReadFile("./testdata/estacoes-gira-1-trimestre-2023.csv")
	s.Require().NoError(err)

	ctx.Request = httptest.NewRequest("GET", "/api/poi?type=gira", nil)
	_, err = s.pois.Import(data, ctx)
	s.Require().NoError(err)

	var points []models.PointOfInterest
	err = s.db.Order("name").Find(&points).Error
	s.Require().NoError(err)

	s.Len(points, 147)
	s.Equal("101 - Alameda dos Oceanos / Rua dos Argonautas", points[0].Name)
	s.Equal(38.756161, points[0].Lat)
	s.Equal(-9.096804, points[0].Lon)
}

func TestPOIController(t *testing.T) {
	acl := access.New()
	registerAllRules(&UserController{}, acl)
	registerAllRules(&POIController{}, acl)
	suite.Run(t, &POIControllerTestSuite{acl: acl})
}
