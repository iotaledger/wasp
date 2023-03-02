package migrations

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestM002CleanupFeePolicy(t *testing.T) {
	oldFeePolicyBin, err := hex.DecodeString("01d2d5150b0c58f98126872ddd41d37254e07deb79e2678fcf6582a1601d84c8a0b4fa8a64e09e0000000001000000640000000a0100000001000000")
	require.NoError(t, err)
	fp, err := m002ConvertFeePolicy(oldFeePolicyBin)
	require.NoError(t, err)
	t.Log(fp)
}
