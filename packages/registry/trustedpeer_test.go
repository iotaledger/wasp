package registry_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/registry"
)

func TestTrustedPeer(t *testing.T) {
	trustedPeersRegistry, err := registry.NewTrustedPeersRegistryImpl("")
	require.NoError(t, err)

	tpList, err := trustedPeersRegistry.TrustedPeers()
	require.NoError(t, err)
	require.Equal(t, 0, len(tpList))

	keyPair1 := cryptolib.NewKeyPair()
	keyPair2 := cryptolib.NewKeyPair()
	_, err = trustedPeersRegistry.TrustPeer("1", keyPair1.GetPublicKey(), "host1:2001")
	require.NoError(t, err)
	_, err = trustedPeersRegistry.TrustPeer("2", keyPair2.GetPublicKey(), "host2a:2002")
	require.NoError(t, err)
	_, err = trustedPeersRegistry.TrustPeer("3", keyPair2.GetPublicKey(), "host2b:2002") // Duplicate entry should be overwritten.
	require.NoError(t, err)

	tpList, err = trustedPeersRegistry.TrustedPeers()
	require.NoError(t, err)
	require.Equal(t, 2, len(tpList))

	_, err = trustedPeersRegistry.DistrustPeer(keyPair1.GetPublicKey())
	require.NoError(t, err)

	tpList, err = trustedPeersRegistry.TrustedPeers()
	require.NoError(t, err)
	require.Equal(t, 1, len(tpList))
	require.True(t, keyPair2.GetPublicKey().Equals(tpList[0].PubKey()))
	require.Equal(t, "host2b:2002", tpList[0].PeeringURL)
}
