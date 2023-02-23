package migrations

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestM001GasPerTokenToRatio32(t *testing.T) {
	oldFeePolicyBin, err := hex.DecodeString("006400000000000000000100000001000000")
	require.NoError(t, err)
	fp, err := m001ConvertFeePolicy(oldFeePolicyBin)
	require.NoError(t, err)
	t.Log(fp)
}
