package registry

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/core/kvstore/mapdb"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
)

func TestTrustedPeer(t *testing.T) {
	// var err error
	log := testlogger.NewLogger(t)
	reg := NewRegistry(log, mapdb.NewMapDB())

	tpList, err := reg.TrustedPeers()
	require.Nil(t, err)
	require.Equal(t, 0, len(tpList))

	keyPair1 := cryptolib.NewKeyPair()
	keyPair2 := cryptolib.NewKeyPair()
	_, err = reg.TrustPeer(keyPair1.GetPublicKey(), "host1:2001")
	require.Nil(t, err)
	_, err = reg.TrustPeer(keyPair2.GetPublicKey(), "host2a:2002")
	require.Nil(t, err)
	_, err = reg.TrustPeer(keyPair2.GetPublicKey(), "host2b:2002") // Duplicate entry should be overwritten.
	require.Nil(t, err)

	tpList, err = reg.TrustedPeers()
	require.Nil(t, err)
	require.Equal(t, 2, len(tpList))

	_, err = reg.DistrustPeer(keyPair1.GetPublicKey())
	require.Nil(t, err)

	tpList, err = reg.TrustedPeers()
	require.Nil(t, err)
	require.Equal(t, 1, len(tpList))
	require.True(t, keyPair2.GetPublicKey().Equals(tpList[0].PubKey))
	require.Equal(t, "host2b:2002", tpList[0].NetID)
}
