package latlon

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDist(t *testing.T) {
	assert.Equal(t, float64(0), Dist(Coords{0, 0}, Coords{0, 0}))
	assert.Equal(t, 5918.185064088765, Dist(Coords{51.5, 0}, Coords{38.8, -77.1}))
}
