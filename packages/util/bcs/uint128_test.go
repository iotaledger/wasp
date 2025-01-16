package bcs_test

import (
	"math/big"
	"testing"

	ref_bcs "github.com/fardream/go-bcs/bcs"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/util/bcs"
)

func TestUint128(t *testing.T) {
	testUint128Codec(t, "10", true)
	testUint128Codec(t, "1770887431076116955186", true)
	testUint128Codec(t, "999999999999999999999999999999999999999999999999999", false)
}

func testUint128Codec(t *testing.T, v string, expectSuccess bool) {
	var bi big.Int
	_, ok := bi.SetString(v, 10)
	require.True(t, ok)

	if expectSuccess {
		refBiEnc := ref_bcs.MustMarshal(lo.Must1(ref_bcs.NewUint128FromBigInt(&bi)))

		bcs.TestCodecAndBytes(t, bi, refBiEnc)
		bcs.TestCodecAndBytes(t, &bi, refBiEnc)
	} else {
		bcs.TestEncodeErr(t, bi)
		bcs.TestEncodeErr(t, &bi)

		_, err := ref_bcs.NewUint128FromBigInt(&bi)
		require.Error(t, err)
	}
}
