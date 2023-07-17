package access

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testtype struct {
	id    string
	owner string
}

func TestAuthorize(t *testing.T) {
	ngn := New()
	ngn.Register(testtype{}, "read", testtype{}, func(ent, res any) bool {
		ttEnt := ent.(testtype)
		ttRes := res.(testtype)
		return ttEnt.id == ttRes.owner
	})
	ngn.Register("", "read,write,delete", testtype{}, func(ent, res any) bool {
		ttRes := res.(testtype)
		return ent == ttRes.owner
	})

	id := "123465-abcdef"
	testcases := []struct {
		ent    any
		action string
		res    any
		result bool
	}{
		{testtype{id: id}, "read", testtype{owner: id}, true},
		{testtype{id: id}, "read", testtype{owner: "not-id"}, false},
		{"aoeusnth", "read", testtype{owner: "aoeusnth"}, true},
		{"aoeusnt", "read", testtype{owner: "aoeusnth"}, false},
		{"aoeusnth", "write", testtype{owner: "aoeusnth"}, true},
		{"aoeusnt", "write", testtype{owner: "aoeusnth"}, false},
		{"aoeusnth", "delete", testtype{owner: "aoeusnth"}, true},
		{"aoeusnt", "delete", testtype{owner: "aoeusnth"}, false},
		{123, "write", "aoeusnth", false},
		{testtype{id: id}, "create", testtype{owner: id}, false},
	}

	for i, tc := range testcases {
		assert.Equal(t,
			tc.result,
			ngn.Authorize(tc.ent, tc.action, tc.res),
			"failed on test case %d", i)
	}
}

func TestKey(t *testing.T) {
	testcases := []struct {
		val any
		exp string
	}{
		{testtype{}, "access.testtype"},
		{&testtype{}, "access.testtype"},
		{1, "int"},
		{"abcde", "string"},
		{false, "bool"},
	}

	for _, tc := range testcases {
		assert.Equal(t, tc.exp, key(tc.val))
	}
}

func TestRuleId(t *testing.T) {
	type val struct {
		ent    any
		action string
		res    any
	}
	testcases := []struct {
		val val
		exp string
	}{
		{val{&testtype{}, "read", 1}, "access.testtype-read-int"},
		{val{"abcde", "write", testtype{}}, "string-write-access.testtype"},
	}

	for _, tc := range testcases {
		assert.Equal(t, tc.exp, ruleId(tc.val.ent, tc.val.action, tc.val.res))
	}
}
