package controllers

import (
	"fmt"
	"testing"

	"bitbucket.org/pensarmais/cycleforlisbon/src/database/models"
	"bitbucket.org/pensarmais/cycleforlisbon/src/server/access"
	"github.com/stretchr/testify/assert"
)

func TestInstitutionACL(t *testing.T) {
	acl := access.New()
	registerAllRules(&InstitutionController{}, acl)

	testCases := []struct {
		ent    models.User
		res    models.Institution
		action string
		exp    bool
	}{
		{
			ent:    models.User{Admin: true},
			res:    models.Institution{},
			action: "create",
			exp:    true,
		},
		{
			ent:    models.User{Admin: false},
			res:    models.Institution{},
			action: "create",
			exp:    false,
		},
		{
			ent:    models.User{Admin: true},
			res:    models.Institution{},
			action: "update",
			exp:    true,
		},
		{
			ent:    models.User{Admin: false},
			res:    models.Institution{},
			action: "update",
			exp:    false,
		},
		{
			ent:    models.User{Admin: true},
			res:    models.Institution{},
			action: "delete",
			exp:    true,
		},
		{
			ent:    models.User{Admin: false},
			res:    models.Institution{},
			action: "delete",
			exp:    false,
		},
		{
			ent:    models.User{Admin: true},
			res:    models.Institution{},
			action: "update-logo",
			exp:    true,
		},
		{
			ent:    models.User{Admin: false},
			res:    models.Institution{},
			action: "update-logo",
			exp:    false,
		},
		{
			ent:    models.User{Admin: true},
			res:    models.Institution{},
			action: "delete-logo",
			exp:    true,
		},
		{
			ent:    models.User{Admin: false},
			res:    models.Institution{},
			action: "delete-logo",
			exp:    false,
		},
	}
	for i, tC := range testCases {
		t.Run(fmt.Sprintf("action: %s, line: %d", tC.action, i), func(t *testing.T) {
			assert.Equal(t, tC.exp, acl.Authorize(tC.ent, tC.action, tC.res))
		})
	}
}
