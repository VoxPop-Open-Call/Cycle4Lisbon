package stringutil

import (
	"regexp"
	"strconv"
)

var floatRegex = regexp.MustCompile(`[+-]?([0-9]*[.])?[0-9]+`)

// AllFloats returns all floats in a string.
func AllFloats(s string) ([]float64, error) {
	raw := floatRegex.FindAllString(s, -1)
	res := make([]float64, len(raw))
	for i, x := range raw {
		v, err := strconv.ParseFloat(x, 64)
		if err != nil {
			return res, err
		}

		res[i] = v
	}

	return res, nil
}
