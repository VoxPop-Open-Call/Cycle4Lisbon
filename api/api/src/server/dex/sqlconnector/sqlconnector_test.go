package sqlconnector

import (
	"context"
	"os"
	"testing"

	"bitbucket.org/pensarmais/cycleforlisbon/src/config"
	"bitbucket.org/pensarmais/cycleforlisbon/src/database"
	"bitbucket.org/pensarmais/cycleforlisbon/src/database/models"
	"bitbucket.org/pensarmais/cycleforlisbon/src/util/password"
	"bitbucket.org/pensarmais/cycleforlisbon/src/util/random"
	"github.com/dexidp/dex/connector"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type SQLConnectorTestSuite struct {
	suite.Suite
	conn  *sqlConnector
	db    *gorm.DB
	users []models.User
}

func (s *SQLConnectorTestSuite) SetupSuite() {
	passwd0 := "password123"
	passwd1 := "aoeusnth"
	hash0, err := password.Hash(passwd0)
	s.NoError(err)
	hash1, err := password.Hash(passwd1)
	s.NoError(err)

	uid0, err := uuid.NewRandom()
	s.NoError(err)
	uid1, err := uuid.NewRandom()
	s.NoError(err)
	users := []models.User{
		{
			Email: random.String(100),
			BaseModel: models.BaseModel{
				ID: uid0,
			},
			Subject:        uid0.String(),
			HashedPassword: hash0,
		},
		{
			Email: random.String(100),
			BaseModel: models.BaseModel{
				ID: uid1,
			},
			Subject:        uid1.String(),
			HashedPassword: hash1,
		},
	}

	err = s.db.CreateInBatches(users, 5).Error
	s.NoError(err)

	s.users = users
}

func (s *SQLConnectorTestSuite) TearDownSuite() {
	err := s.db.Delete(&models.User{}, "id = ?", s.users[0].ID).Error
	s.NoError(err)
	err = s.db.Delete(&models.User{}, "id = ?", s.users[1].ID).Error
	s.NoError(err)
}

func (s *SQLConnectorTestSuite) TestLogin() {
	ctx := context.Background()
	scopes := connector.Scopes{}
	ident, ok, err := s.conn.Login(ctx, scopes, s.users[0].Email, "password123")
	s.NotEmpty(ident)
	s.True(ok)
	s.NoError(err)

	s.Equal(connector.Identity{
		UserID:            s.users[0].ID.String(),
		Username:          s.users[0].ID.String(),
		PreferredUsername: "",
		Email:             s.users[0].Email,
		EmailVerified:     false,
	}, ident)

	ident, ok, err = s.conn.Login(ctx, scopes, s.users[1].Email, "aoeusnth")
	s.NotEmpty(ident)
	s.True(ok)
	s.NoError(err)

	s.Equal(connector.Identity{
		UserID:            s.users[1].ID.String(),
		Username:          s.users[1].ID.String(),
		PreferredUsername: "",
		Email:             s.users[1].Email,
		EmailVerified:     false,
	}, ident)
}

func (s *SQLConnectorTestSuite) TestLoginWrongPassword() {
	ctx := context.Background()
	scopes := connector.Scopes{}
	ident, ok, err := s.conn.Login(ctx, scopes, s.users[0].Email, "qwerty")
	s.Empty(ident)
	s.False(ok)
	s.NoError(err)

	ident, ok, err = s.conn.Login(ctx, scopes, s.users[0].Email, "")
	s.Empty(ident)
	s.False(ok)
	s.NoError(err)
}

func (s *SQLConnectorTestSuite) TestRefresh() {
	ctx := context.Background()
	scopes := connector.Scopes{
		OfflineAccess: true,
	}
	ident, ok, err := s.conn.Login(ctx, scopes, s.users[0].Email, "password123")
	s.NotEmpty(ident)
	s.True(ok)
	s.NoError(err)

	ctx = context.Background()
	newIdent, err := s.conn.Refresh(ctx, scopes, ident)
	s.NotEmpty(newIdent)
	s.NoError(err)

	s.Equal(connector.Identity{
		UserID:            s.users[0].ID.String(),
		Username:          s.users[0].ID.String(),
		PreferredUsername: "",
		Email:             s.users[0].Email,
		EmailVerified:     false,
		ConnectorData:     ident.ConnectorData,
	}, newIdent)
}

func TestSQLConnector(t *testing.T) {
	conf, err := config.Load("../../../../.env")
	require.NoError(t, err)

	lgr := &logrus.Logger{
		Out:   os.Stdout,
		Level: logrus.ErrorLevel,
	}
	connConfig := Config{
		DSN: conf.DbDsn(),
	}
	conn, err := connConfig.Open("custom", lgr)
	require.NoError(t, err)
	sqlConn, ok := conn.(*sqlConnector)
	require.True(t, ok)

	db, err := database.Init(conf.DbDsn())
	require.NoError(t, err)
	db.Logger = logger.Default.LogMode(logger.Silent)

	suite.Run(t, &SQLConnectorTestSuite{
		conn: sqlConn,
		db:   db,
	})
}
