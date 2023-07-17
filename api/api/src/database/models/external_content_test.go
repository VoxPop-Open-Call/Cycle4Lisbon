package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExternalContentScan(t *testing.T) {
	for i, tc := range []struct {
		src any
		err string
		exp string
	}{
		{"approved", "", "approved"},
		{"pending", "", "pending"},
		{"rejected", "", "rejected"},
		{"", "", ""},
		{"invalid", "invalid value for ExternalContentState: invalid", ""},
		{[]byte("approved"), "", "approved"},
		{[]byte("rejected"), "", "rejected"},
		{[]byte("pending"), "", "pending"},
		{[]byte(""), "", ""},
		{[]byte("invalid"), "invalid value for ExternalContentState: invalid", ""},
	} {
		ms := new(ExternalContentState)
		err := ms.Scan(tc.src)
		if tc.err == "" {
			assert.NoError(t, err, "failed test case %d", i)
		} else {
			assert.EqualError(t, err, tc.err, "failed test case %d", i)
		}
		assert.Equal(t, tc.exp, string(*ms), "failed test case %d", i)
	}
}

func TestExternalContentValue(t *testing.T) {
	for i, tc := range []struct {
		ms  ExternalContentState
		err string
		exp string
	}{
		{"approved", "", "approved"},
		{"pending", "", "pending"},
		{"rejected", "", "rejected"},
		{"", "", ""},
		{"invalid", "invalid value for ExternalContentState: invalid", ""},
	} {
		val, err := tc.ms.Value()
		if tc.err == "" {
			assert.NoError(t, err, "failed test case %d", i)
		} else {
			assert.EqualError(t, err, tc.err, "failed test case %d", i)
		}
		assert.Equal(t, tc.exp, val, "failed test case %d", i)
	}
}
