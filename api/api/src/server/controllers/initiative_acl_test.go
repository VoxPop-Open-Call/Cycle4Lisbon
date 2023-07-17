package controllers

import (
	"testing"

	"bitbucket.org/pensarmais/cycleforlisbon/src/database/models"
	"bitbucket.org/pensarmais/cycleforlisbon/src/server/access"
	"github.com/stretchr/testify/assert"
)

func TestInitiativeAcl(t *testing.T) {
	acl := access.New()
	registerAllRules(&InitiativeController{}, acl)

	testcases := []struct {
		ent, res any
		action   string
		exp      bool
	}{
		{
			ent:    models.User{Admin: true},
			res:    models.Initiative{},
			action: "create",
			exp:    true,
		},
		{
			ent:    models.User{Admin: false},
			res:    models.Initiative{},
			action: "create",
			exp:    false,
		},
		{
			ent:    models.User{Admin: true},
			res:    models.Initiative{},
			action: "update",
			exp:    true,
		},
		{
			ent:    models.User{Admin: false},
			res:    models.Initiative{},
			action: "update",
			exp:    false,
		},
		{
			ent:    models.User{Admin: true},
			res:    models.Initiative{},
			action: "change-state",
			exp:    true,
		},
		{
			ent:    models.User{Admin: false},
			res:    models.Initiative{},
			action: "change-state",
			exp:    false,
		},
		{
			ent:    models.User{Admin: true},
			res:    models.Initiative{},
			action: "delete",
			exp:    true,
		},
		{
			ent:    models.User{Admin: false},
			res:    models.Initiative{},
			action: "delete",
			exp:    false,
		},
		{
			ent:    models.User{Admin: true},
			res:    models.Initiative{},
			action: "update-img",
			exp:    true,
		},
		{
			ent:    models.User{Admin: false},
			res:    models.Initiative{},
			action: "update-img",
			exp:    false,
		},
		{
			ent:    models.User{Admin: true},
			res:    models.Initiative{},
			action: "delete-img",
			exp:    true,
		},
		{
			ent:    models.User{Admin: false},
			res:    models.Initiative{},
			action: "delete-img",
			exp:    false,
		},
	}

	for i, tc := range testcases {
		assert.Equal(
			t,
			tc.exp,
			acl.Authorize(tc.ent, tc.action, tc.res),
			"failed for test case %d", i,
		)
	}
}
