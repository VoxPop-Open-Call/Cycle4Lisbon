package stringutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAllFloats(t *testing.T) {
	testCases := []struct {
		desc string
		exp  []float64
	}{
		{
			desc: "123",
			exp:  []float64{123},
		},
		{
			desc: "123 12.34",
			exp:  []float64{123, 12.34},
		},
		{
			desc: ".456 123 12.34",
			exp:  []float64{0.456, 123, 12.34},
		},
		{
			desc: "{'coordinates': [-9.14207, 38.74873], 'type': 'Point'}",
			exp:  []float64{-9.14207, 38.74873},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			res, err := AllFloats(tC.desc)
			assert.NoError(t, err)
			assert.Equal(t, tC.exp, res)
		})
	}
}
