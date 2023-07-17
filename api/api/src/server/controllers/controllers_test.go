package controllers

import (
	"log"
	"net/http/httptest"
	"os"
	"testing"

	"bitbucket.org/pensarmais/cycleforlisbon/src/config"
	"bitbucket.org/pensarmais/cycleforlisbon/src/database"
	"bitbucket.org/pensarmais/cycleforlisbon/src/database/models"
	"bitbucket.org/pensarmais/cycleforlisbon/src/server/middleware"
	"bitbucket.org/pensarmais/cycleforlisbon/src/util/random"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var testDb *gorm.DB

type MockPresigner struct {
	mock.Mock
}

// createRandomUser creates a new user, returns it, and also returns a gin test
// context with the token claims set to the new user.
func createRandomUser(users *UserController) (models.User, *gin.Context, error) {
	params := CreateUserParams{
		Name:     random.String(5),
		Email:    random.String(100) + "@test.org",
		Password: "pass@word",
	}

	user, err := users.Create(params, nil)

	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Set(middleware.TokenClaimsKey, middleware.Claims{
		Sub: user.ID.String(),
	})

	return user, ctx, err
}

// createRandomAdmin is like createRandomUser, except the created user has
// admin privileges.
func createRandomAdmin(users *UserController) (models.User, *gin.Context, error) {
	user := models.User{
		Email: random.String(100) + "@test.org",
		Admin: true,
		Profile: models.Profile{
			Name: random.String(50),
		},
	}
	err := users.db.Create(&user).Error

	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Set(middleware.TokenClaimsKey, middleware.Claims{
		Sub: user.ID.String(),
	})

	return user, ctx, err
}

func TestOrderBy(t *testing.T) {
	for i, tc := range []struct {
		val, exp string
	}{
		{"id asc", "id asc"},
		{"id,name asc", "id,name asc"},
		{"id, name asc", "id, name asc"},
		{"createdAt", "created_at"},
		{"createdAt, updatedAt, email asc", "created_at, updated_at, email asc"},
		{"articleURL", "article_url"},
		{"FCMToken", "fcm_token"},
	} {
		assert.Equal(t, tc.exp, orderBy(tc.val).ToSnakeCase(),
			"failed on test %d: %s", i, tc.val)
	}
}

func TestOrderByValidationRegex(t *testing.T) {
	for i, tc := range []struct {
		val string
		exp bool
	}{
		{"id asc", true},
		{"id, name asc", true},
		{"id, createdAt desc", true},
		{"id, name, data123 desc", true},
		{"id,name asc", true},
		{"id,name desc", true},
		{"id,name desc;drop table users", false},
		{"id, name desc, (select 1)", false},
		{"id, name, email, desc", false},
	} {
		assert.Equal(t, tc.exp,
			orderByValidationRegex.MatchString(tc.val),
			"failed on test case %d", i)
	}
}

// Connect to the database to test the controllers.
// Each controller test suite should run inside a transaction, to prevent side effects.
func TestMain(m *testing.M) {
	config, err := config.Load("../../../.env")
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}
	gin.SetMode(gin.TestMode)

	// Register custom validation tags
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		for k, f := range customValidations {
			v.RegisterValidation(k, f)
		}
	}

	testDb, err = database.Init(config.DbDsn())
	if err != nil {
		log.Fatalf("error connecting to the database: %v", err)
	}

	testDb.Logger = logger.Default.LogMode(logger.Silent)

	os.Exit(m.Run())
}
