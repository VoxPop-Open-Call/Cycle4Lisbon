package controllers

import (
	"testing"

	"bitbucket.org/pensarmais/cycleforlisbon/src/database/models"
	"bitbucket.org/pensarmais/cycleforlisbon/src/server/access"
	"github.com/stretchr/testify/assert"
)

func TestExternalContentAcl(t *testing.T) {
	acl := access.New()
	registerAllRules(&ExternalContentController{}, acl)

	testcases := []struct {
		ent, res any
		action   string
		exp      bool
	}{
		{
			ent:    models.User{Admin: true},
			res:    models.ExternalContent{},
			action: "change-state",
			exp:    true,
		},
		{
			ent:    models.User{Admin: false},
			res:    models.ExternalContent{},
			action: "change-state",
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
