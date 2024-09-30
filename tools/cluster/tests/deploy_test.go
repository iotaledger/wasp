package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/clients/apiextensions"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/corecontracts"
	"github.com/iotaledger/wasp/packages/vm/core/inccounter"
	"github.com/iotaledger/wasp/packages/vm/core/root"
)

// executed in cluster_test.go
func testDeployChain(t *testing.T, env *ChainEnv) {
	chainID, chainOwnerID := env.getChainInfo()
	require.EqualValues(t, chainID, env.Chain.ChainID)
	require.EqualValues(t, chainOwnerID, isc.NewAddressAgentID(env.Chain.OriginatorAddress()))
	t.Logf("--- chainID: %s", chainID.String())
	t.Logf("--- chainOwnerID: %s", chainOwnerID.String())

	env.checkCoreContracts()
	env.checkRootsOutside()
	for _, i := range env.Chain.CommitteeNodes {
		blockIndex, err := env.Chain.BlockIndex(i)
		require.NoError(t, err)
		require.Greater(t, blockIndex, uint32(1))

		contractRegistry, err := env.Chain.ContractRegistry(i)
		require.NoError(t, err)
		require.EqualValues(t, len(corecontracts.All), len(contractRegistry))
	}
}

// executed in cluster_test.go
func testIncCounterIsDeployed(t *testing.T, env *ChainEnv) {
	// test calling root.FuncFindContractByName view function using client
	ret, err := apiextensions.CallView(
		context.Background(),
		env.Chain.Cluster.WaspClient(),
		env.Chain.ChainID.String(),
		apiclient.ContractCallViewRequest{
			ContractHName: root.Contract.Hname().String(),
			FunctionHName: root.ViewFindContract.Hname().String(),
			Arguments: apiextensions.DictToAPIJsonDict(dict.Dict{
				root.ParamHname: inccounter.Contract.Hname().Bytes(),
			}),
		})

	require.NoError(t, err)
	recb := ret.Get(root.ParamContractRecData)
	_, err = root.ContractRecordFromBytes(recb)
	require.NoError(t, err)
}
