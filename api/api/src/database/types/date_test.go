package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDateScan(t *testing.T) {
	src, err := time.Parse(DateFormat, "2023-05-30")

	var date Date
	err = date.Scan(src)
	assert.NoError(t, err)
	assert.Equal(t, "2023-05-30", string(date))
}

func TestDateJsonEncoding(t *testing.T) {
	date := Date("2000-01-30")
	b, err := date.MarshalJSON()
	require.NoError(t, err)

	assert.Equal(t, "\"2000-01-30\"", string(b))

	var res Date
	err = res.UnmarshalJSON(b)
	require.NoError(t, err)
	require.Equal(t, date, res)
}
