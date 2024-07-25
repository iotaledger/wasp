package isc

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/util/rwutil"
)

func TestChainIDSerialization(t *testing.T) {
	chainID := RandomChainID()
	rwutil.ReadWriteTest(t, &chainID, new(ChainID))
	rwutil.BytesTest(t, chainID, ChainIDFromBytes)
	rwutil.StringTest(t, chainID, ChainIDFromString)
}

func TestIncorrectPrefix(t *testing.T) {
	chainID := "rms1prxunz807j39nmhzy3gre4hwdlzvdjyrkfn59d27x6xh426y8ajt205mh9g"
	_, err := ChainIDFromString(chainID)

	require.ErrorContains(t, err, "invalid network prefix: rms")
}
