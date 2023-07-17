package password

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPassword(t *testing.T) {
	passwords := []string{
		"UmWfZXeZ#cn }J#^~kC",
		"NExSY~U#MijWmEfVF",
		"#CyTvm(^@aeL)L&GyBt]]}",
		"p]Se*MPnqZGZq}*JGajj",
		"kC%tfN*Sq$pZz@PT)en^ju",
		"FUl~CvW]ohT",
		"hU&cW%A{&@[spl*He!y",
		"]kS$Vv{luADw}oAS",
		"fnGdPxvk%[S})[bYM",
		"&EcwGdurCIJXbGFSt}",
	}

	for _, pwd := range passwords {
		hash, err := Hash(pwd)
		require.NoError(t, err)
		require.NotZero(t, hash)

		ok := Check(pwd, hash)
		require.True(t, ok)

		ok = Check("password123", hash)
		require.False(t, ok)
	}
}

func TestEmptyPassword(t *testing.T) {
	require.False(t, Check("", ""))
	require.False(t, Check("password1234", ""))
}
