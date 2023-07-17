package database

import (
	"testing"

	"bitbucket.org/pensarmais/cycleforlisbon/src/config"
	"bitbucket.org/pensarmais/cycleforlisbon/src/database/models"
	"bitbucket.org/pensarmais/cycleforlisbon/src/util/random"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestInsertAdminUser(t *testing.T) {
	config, err := config.Load("../../.env")
	require.NoError(t, err)

	db, err := Init(config.DbDsn())
	require.NoError(t, err)
	db.Logger = logger.Default.LogMode(logger.Silent)
	tx := db.Begin()
	tx.Exec("DELETE FROM users")

	// ------------------------ //
	// Creates a new admin user //
	// ------------------------ //
	email := random.String(50) + "@admin.test"
	passwd := "supersafepw0000"

	err = CreateAdminUser(email, passwd, tx)
	require.NoError(t, err)

	var user models.User
	err = tx.First(&user, "email = ?", email).Error
	require.NoError(t, err)
	require.NotEmpty(t, user)

	require.NotEmpty(t, user.ID.String())
	require.Equal(t, user.ID.String(), user.Subject)
	require.True(t, user.Admin)
	require.True(t, user.Verified)

	// ----------------------------------------------------- //
	// Doesn't create another admin if there are any already //
	// ----------------------------------------------------- //
	email = random.String(25) + "@admin.test"
	err = CreateAdminUser(email, passwd, tx)
	require.NoError(t, err)

	user = models.User{}
	err = tx.First(&user, "email = ?", email).Error
	require.EqualError(t, err, gorm.ErrRecordNotFound.Error())
	require.Empty(t, user)

	tx.Rollback()
}
