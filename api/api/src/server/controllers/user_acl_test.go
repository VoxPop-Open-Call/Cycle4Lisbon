package controllers

import (
	"testing"

	"bitbucket.org/pensarmais/cycleforlisbon/src/database/models"
	"bitbucket.org/pensarmais/cycleforlisbon/src/server/access"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUsersAcl(t *testing.T) {
	acl := access.New()
	registerAllRules(&UserController{}, acl)

	uid1, err := uuid.NewRandom()
	require.NoError(t, err)
	uid2, err := uuid.NewRandom()
	require.NoError(t, err)

	for i, tc := range []struct {
		ent, res models.User
		action   string
		exp      bool
	}{
		{
			ent:    models.User{BaseModel: models.BaseModel{ID: uid1}},
			res:    models.User{BaseModel: models.BaseModel{ID: uid1}},
			action: "update",
			exp:    true,
		},
		{
			ent:    models.User{BaseModel: models.BaseModel{ID: uid1}},
			res:    models.User{BaseModel: models.BaseModel{ID: uid2}},
			action: "update",
			exp:    false,
		},
		{
			ent:    models.User{BaseModel: models.BaseModel{ID: uid1}},
			res:    models.User{BaseModel: models.BaseModel{ID: uid1}},
			action: "delete",
			exp:    true,
		},
		{
			ent:    models.User{BaseModel: models.BaseModel{ID: uid1}},
			res:    models.User{BaseModel: models.BaseModel{ID: uid2}},
			action: "delete",
			exp:    false,
		},
		{
			ent:    models.User{Admin: true},
			res:    models.User{BaseModel: models.BaseModel{ID: uid2}},
			action: "delete",
			exp:    true,
		},
		{
			ent:    models.User{BaseModel: models.BaseModel{ID: uid1}},
			res:    models.User{BaseModel: models.BaseModel{ID: uid1}},
			action: "update-picture",
			exp:    true,
		},
		{
			ent:    models.User{BaseModel: models.BaseModel{ID: uid1}},
			res:    models.User{BaseModel: models.BaseModel{ID: uid2}},
			action: "update-picture",
			exp:    false,
		},
		{
			ent:    models.User{BaseModel: models.BaseModel{ID: uid1}},
			res:    models.User{BaseModel: models.BaseModel{ID: uid1}},
			action: "delete-picture",
			exp:    true,
		},
		{
			ent:    models.User{BaseModel: models.BaseModel{ID: uid1}},
			res:    models.User{BaseModel: models.BaseModel{ID: uid2}},
			action: "delete-picture",
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
