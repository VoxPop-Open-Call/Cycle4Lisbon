package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetFunc(t *testing.T) {
	os.Setenv("key-0", "abcd")
	m := make(map[string]string)
	m["key-1"] = "1234"

	var tests = []struct {
		key string
		def string
		exp string
	}{
		{"key-0", "efhd", "abcd"},
		{"key-1", "5678", "1234"},
		{"key-2", "!@#$", "!@#$"},
	}

	for _, test := range tests {
		res := get(m, test.key, test.def)
		require.Equal(t, test.exp, res)
	}
}

func TestLoad(t *testing.T) {
	os.Setenv("GIN_MODE", "release")

	conf, err := Load("")
	require.NotEmpty(t, conf)
	require.NoError(t, err)

	assert.Equal(t, "release", conf.GIN_MODE)
	assert.Equal(t, "localhost", conf.DB_HOST)
	assert.Equal(t, uint16(5432), conf.DEX_DB_PORT)
}
