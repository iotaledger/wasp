package util_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

func TestRatioSerialization(t *testing.T) {
	ratio1 := util.Ratio32{
		A: 123,
		B: 246,
	}

	b := ratio1.Bytes()
	ratio2, err := util.Ratio32FromBytes(b)
	require.NoError(t, err)
	require.Equal(t, ratio1, ratio2)
	s := ratio1.String()
	ratio3, err := util.Ratio32FromString(s)
	require.NoError(t, err)
	require.Equal(t, ratio2, ratio3)

	rwutil.ReadWriteTest(t, &ratio1, new(util.Ratio32))
}
