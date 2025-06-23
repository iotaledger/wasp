package tests

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/core/corecontracts"
)

// executed in cluster_test.go
func (e *ChainEnv) testDeployChain(t *testing.T) {
	chainID, chainAdmin := e.getChainInfo()
	require.EqualValues(t, chainID, e.Chain.ChainID)
	require.EqualValues(t, chainAdmin, isc.NewAddressAgentID(e.Chain.OriginatorAddress()))
	t.Logf("--- chainID: %s", chainID.String())
	t.Logf("--- chainAdmin: %s", chainAdmin.String())

	e.checkCoreContracts()
	e.checkRootsOutside()
	for _, i := range e.Chain.CommitteeNodes {
		blockIndex, err := e.Chain.BlockIndex(i)
		require.NoError(t, err)
		require.Greater(t, blockIndex, uint32(1))

		contractRegistry, err := e.Chain.ContractRegistry(i)
		require.NoError(t, err)
		require.EqualValues(t, len(corecontracts.All), len(contractRegistry))
	}
}
