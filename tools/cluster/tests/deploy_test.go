package tests

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/clients/apiextensions"
	"github.com/iotaledger/wasp/clients/chainclient"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/corecontracts"
	"github.com/iotaledger/wasp/packages/vm/core/root"
)

// executed in cluster_test.go
func testDeployChain(t *testing.T, env *ChainEnv) {
	chainID, chainOwnerID := env.getChainInfo()
	require.EqualValues(t, chainID, env.Chain.ChainID)
	require.EqualValues(t, chainOwnerID, isc.NewAgentID(env.Chain.OriginatorAddress()))
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
func testDeployContractOnly(t *testing.T, env *ChainEnv) {
	env.deployNativeIncCounterSC()

	// test calling root.FuncFindContractByName view function using client
	ret, err := apiextensions.CallView(context.Background(), env.Chain.Cluster.WaspClient(), apiclient.ContractCallViewRequest{
		ChainId:       env.Chain.ChainID.String(),
		ContractHName: root.Contract.Hname().String(),
		FunctionHName: root.ViewFindContract.Hname().String(),
		Arguments: apiextensions.DictToAPIJsonDict(dict.Dict{
			root.ParamHname: isc.Hn(nativeIncCounterSCName).Bytes(),
		}),
	})

	require.NoError(t, err)
	recb, err := ret.Get(root.ParamContractRecData)
	require.NoError(t, err)
	rec, err := root.ContractRecordFromBytes(recb)
	require.NoError(t, err)
	require.EqualValues(t, "testing contract deployment with inccounter", rec.Description)
}

// executed in cluster_test.go
func testDeployContractAndSpawn(t *testing.T, env *ChainEnv) {
	env.deployNativeIncCounterSC()

	hname := isc.Hn(nativeIncCounterSCName)

	nameNew := "spawnedContract"
	dscrNew := "spawned contract it is"
	hnameNew := isc.Hn(nameNew)
	// send 'spawn' request to the SC which was just deployed
	par := chainclient.NewPostRequestParams(
		inccounter.VarName, nameNew,
		inccounter.VarDescription, dscrNew,
	).WithBaseTokens(100)
	tx, err := env.Chain.OriginatorClient().Post1Request(hname, inccounter.FuncSpawn.Hname(), *par)
	require.NoError(t, err)

	receipts, err := env.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(env.Chain.ChainID, tx, 30*time.Second)
	require.NoError(t, err)
	require.Len(t, receipts, 1)

	env.checkCoreContracts()
	for _, i := range env.Chain.CommitteeNodes {
		blockIndex, err := env.Chain.BlockIndex(i)
		require.NoError(t, err)
		require.Greater(t, blockIndex, uint32(2))

		contractRegistry, err := env.Chain.ContractRegistry(i)
		require.NoError(t, err)
		require.EqualValues(t, len(corecontracts.All)+2, len(contractRegistry))

		cr := contractRegistry[hnameNew]
		require.EqualValues(t, dscrNew, cr.Description)
		require.EqualValues(t, nameNew, cr.Name)

		counterValue, err := env.Chain.GetCounterValue(hname, i)
		require.NoError(t, err)
		require.EqualValues(t, 42, counterValue)
	}
}
