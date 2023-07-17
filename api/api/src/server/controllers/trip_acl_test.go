package controllers

import (
	"testing"

	"bitbucket.org/pensarmais/cycleforlisbon/src/database/models"
	"bitbucket.org/pensarmais/cycleforlisbon/src/server/access"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTripsAcl(t *testing.T) {
	acl := access.New()
	registerAllRules(&TripController{}, acl)

	uid1, err := uuid.NewRandom()
	require.NoError(t, err)
	uid2, err := uuid.NewRandom()
	require.NoError(t, err)

	for i, tc := range []struct {
		ent    models.User
		res    models.Trip
		action string
		exp    bool
	}{
		{
			ent:    models.User{BaseModel: models.BaseModel{ID: uid1}},
			res:    models.Trip{UserID: uid1},
			action: "get",
			exp:    true,
		},
		{
			ent:    models.User{BaseModel: models.BaseModel{ID: uid1}},
			res:    models.Trip{UserID: uid2},
			action: "get",
			exp:    false,
		},
	} {
		assert.Equal(
			t,
			tc.exp,
			acl.Authorize(tc.ent, tc.action, tc.res),
			"failed on test %d", i,
		)
	}
}
