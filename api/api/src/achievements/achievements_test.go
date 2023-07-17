package achievements

import (
	"log"
	"os"
	"testing"
	"time"

	"bitbucket.org/pensarmais/cycleforlisbon/src/config"
	"bitbucket.org/pensarmais/cycleforlisbon/src/database"
	"bitbucket.org/pensarmais/cycleforlisbon/src/database/models"
	"bitbucket.org/pensarmais/cycleforlisbon/src/database/query"
	"bitbucket.org/pensarmais/cycleforlisbon/src/util/random"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

var testDb *gorm.DB

func TestAchievements(t *testing.T) {
	db := testDb.Begin()
	defer db.Rollback()

	achievements, err := New(db, query.Achievements)
	require.NoError(t, err)
	require.NotEmpty(t, achievements)

	user := models.User{
		BaseModel: models.BaseModel{ID: uuid.New()},
		Subject:   random.String(32),
	}
	db.Create(&user)

	newAchs, err := achievements.Update(user.ID, State{
		Rides:       3,
		Distance:    30,
		Initiatives: 4,
		Credits:     5,
	})
	require.NoError(t, err)
	assert.Len(t, newAchs, 4)

	now := time.Now()

	newAchs[0].AchievedAt = &now
	assert.Equal(t, models.UserAchievement{
		UserID:          user.ID,
		AchievementCode: "rides-beginner",
		Achieved:        true,
		AchievedAt:      &now,
		Completion:      1,
	}, newAchs[0])

	newAchs[1].AchievedAt = &now
	assert.Equal(t, models.UserAchievement{
		UserID:          user.ID,
		AchievementCode: "dst-training-wheels",
		Achieved:        true,
		AchievedAt:      &now,
		Completion:      1,
	}, newAchs[1])

	newAchs, err = achievements.Update(user.ID, State{
		Rides:       100,
		Distance:    30,
		Initiatives: 4,
		Credits:     5,
	})
	require.NoError(t, err)
	assert.Len(t, newAchs, 2)

	newAchs[0].AchievedAt = &now
	assert.Equal(t, models.UserAchievement{
		UserID:          user.ID,
		AchievementCode: "rides-traveler",
		Achieved:        true,
		AchievedAt:      &now,
		Completion:      1,
	}, newAchs[0])
	newAchs[1].AchievedAt = &now

	assert.Equal(t, models.UserAchievement{
		UserID:          user.ID,
		AchievementCode: "rides-pro",
		Achieved:        true,
		AchievedAt:      &now,
		Completion:      1,
	}, newAchs[1])
}

func TestMain(m *testing.M) {
	config, err := config.Load("../../.env")
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	testDb, err = database.Init(config.DbDsn())
	if err != nil {
		log.Fatalf("error connecting to the database: %v", err)
	}

	os.Exit(m.Run())
}
