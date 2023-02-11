package registry

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/cryptolib"
)

func TestTrustedPeer(t *testing.T) {
	// var err error
	trustedPeersRegistry, err := NewTrustedPeersRegistryImpl("")
	require.Nil(t, err)

	tpList, err := trustedPeersRegistry.TrustedPeers()
	require.Nil(t, err)
	require.Equal(t, 0, len(tpList))

	keyPair1 := cryptolib.NewKeyPair()
	keyPair2 := cryptolib.NewKeyPair()
	_, err = trustedPeersRegistry.TrustPeer("1", keyPair1.GetPublicKey(), "host1:2001")
	require.Nil(t, err)
	_, err = trustedPeersRegistry.TrustPeer("2", keyPair2.GetPublicKey(), "host2a:2002")
	require.Nil(t, err)
	_, err = trustedPeersRegistry.TrustPeer("3", keyPair2.GetPublicKey(), "host2b:2002") // Duplicate entry should be overwritten.
	require.Nil(t, err)

	tpList, err = trustedPeersRegistry.TrustedPeers()
	require.Nil(t, err)
	require.Equal(t, 2, len(tpList))

	_, err = trustedPeersRegistry.DistrustPeer(keyPair1.GetPublicKey())
	require.Nil(t, err)

	tpList, err = trustedPeersRegistry.TrustedPeers()
	require.Nil(t, err)
	require.Equal(t, 1, len(tpList))
	require.True(t, keyPair2.GetPublicKey().Equals(tpList[0].PubKey()))
	require.Equal(t, "host2b:2002", tpList[0].PeeringURL)
}
