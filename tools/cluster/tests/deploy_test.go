package tests

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/isc"
)

// executed in cluster_test.go
func testDeployChain(t *testing.T, env *ChainEnv) {
	chainID, chainAdmin := env.getChainInfo()
	require.EqualValues(t, chainID, env.Chain.ChainID)
	require.EqualValues(t, chainAdmin, isc.NewAddressAgentID(env.Chain.OriginatorAddress()))
	t.Logf("--- chainID: %s", chainID.String())
	t.Logf("--- chainAdmin: %s", chainAdmin.String())

	env.checkCoreContracts()
	env.checkRootsOutside()
	for _, i := range env.Chain.CommitteeNodes {
		blockIndex, err := env.Chain.BlockIndex(i)
		require.NoError(t, err)
		require.Greater(t, blockIndex, uint32(1))
	}
}
