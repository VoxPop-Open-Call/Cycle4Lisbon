package gobutil

import (
	"sync"
	"testing"

	"bitbucket.org/pensarmais/cycleforlisbon/src/util/random"
	"github.com/stretchr/testify/require"
)

type testtype struct {
	A string
	B int
	C bool
}

func randomTestCase() testtype {
	return testtype{
		A: random.String(500),
		B: random.Int(0, 10000000),
		C: random.Bool(),
	}
}

func TestConcurrentAccess(t *testing.T) {
	testcases := make([]testtype, 100)
	for i := 0; i < len(testcases); i++ {
		testcases[i] = randomTestCase()
	}

	codec := NewGobCodec[testtype]()

	wg := &sync.WaitGroup{}
	for _, tc := range testcases {
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(tc testtype) {
				encoded, err := codec.Encode(tc)
				require.NotEmpty(t, encoded)
				require.NoError(t, err)

				decoded, err := codec.Decode(encoded)
				require.NotEmpty(t, decoded)
				require.NoError(t, err)

				require.Equal(t, tc, decoded)
				wg.Done()
			}(tc)
		}
	}

	wg.Wait()
}
