package registry

import (
	"testing"

	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/stretchr/testify/require"
)

func TestTrustedPeer(t *testing.T) {
	// var err error
	log := testlogger.NewLogger(t)
	reg := NewRegistry(log, mapdb.NewMapDB())

	tpList, err := reg.TrustedPeers()
	require.Nil(t, err)
	require.Equal(t, 0, len(tpList))

	keyPair1 := ed25519.GenerateKeyPair()
	keyPair2 := ed25519.GenerateKeyPair()
	_, err = reg.TrustPeer(keyPair1.PublicKey, "host1:2001")
	require.Nil(t, err)
	_, err = reg.TrustPeer(keyPair2.PublicKey, "host2a:2002")
	require.Nil(t, err)
	_, err = reg.TrustPeer(keyPair2.PublicKey, "host2b:2002") // Duplicate entry should be overwritten.
	require.Nil(t, err)

	tpList, err = reg.TrustedPeers()
	require.Nil(t, err)
	require.Equal(t, 2, len(tpList))

	_, err = reg.DistrustPeer(keyPair1.PublicKey)
	require.Nil(t, err)

	tpList, err = reg.TrustedPeers()
	require.Nil(t, err)
	require.Equal(t, 1, len(tpList))
	require.Equal(t, keyPair2.PublicKey, tpList[0].PubKey)
	require.Equal(t, "host2b:2002", tpList[0].NetID)
}
