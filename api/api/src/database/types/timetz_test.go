package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTimeTZScan(t *testing.T) {
	for i, tc := range []struct {
		val string
		exp string
	}{
		{"04:30:00+02:00", "04:30+02:00"},
		{"12:00:00-02", "12:00-02:00"},
		{"04:30:00+02", "04:30+02:00"},
		{"18:00:00+00", "18:00Z"},
		{"18:00:00Z", "18:00Z"},
	} {
		var ttz TimeTZ
		err := ttz.Scan(tc.val)
		assert.NoError(t, err, "failed on test %d", i)
		assert.Equal(t, tc.exp, ttz.String(), "failed on test %d", i)
	}
}

func TestTimeTZJsonEncoding(t *testing.T) {
	for i, tc := range []string{
		"04:30+02:00",
		"12:00-02:00",
		"23:18+02:00",
		"18:00Z",
	} {
		ttz := TimeTZ(tc)
		b, err := ttz.MarshalJSON()
		assert.NoError(t, err, "failed on test %d", i)

		assert.Equal(t, "\""+tc+"\"", string(b), "failed on test %d", i)

		var res TimeTZ
		err = res.UnmarshalJSON(b)
		assert.NoError(t, err, "failed on test %d", i)

		assert.Equal(t, ttz, res, "failed on test %d", i)
	}
}
