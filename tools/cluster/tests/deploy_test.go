package tests

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/corecontracts"
	"github.com/iotaledger/wasp/packages/vm/core/root"
)

func TestDeployChain(t *testing.T) {
	e := setupWithNoChain(t)

	counter1, err := e.Clu.StartMessageCounter(map[string]int{
		"dismissed_chain": 0,
		"state":           0, // TODO: Was 2, is it meaningful to check this?
		"request_out":     6, // Init + (N=4)*AddAccessNodes + ChangeAccessNodes
	})
	require.NoError(t, err)
	defer counter1.Close()

	chain, err := e.Clu.DeployDefaultChain()
	require.NoError(t, err)

	chEnv := newChainEnv(t, e.Clu, chain)

	if !counter1.WaitUntilExpectationsMet() {
		t.Fatal()
	}
	chainID, chainOwnerID := chEnv.getChainInfo()
	require.EqualValues(t, chainID, chain.ChainID)
	require.EqualValues(t, chainOwnerID, isc.NewAgentID(chain.OriginatorAddress()))
	t.Logf("--- chainID: %s", chainID.String())
	t.Logf("--- chainOwnerID: %s", chainOwnerID.String())

	chEnv.checkCoreContracts()
	chEnv.checkRootsOutside()
	for _, i := range chain.CommitteeNodes {
		blockIndex, err := chain.BlockIndex(i)
		require.NoError(t, err)
		require.Greater(t, blockIndex, uint32(1))

		contractRegistry, err := chain.ContractRegistry(i)
		require.NoError(t, err)
		require.EqualValues(t, len(corecontracts.All), len(contractRegistry))
	}
}

func TestDeployContractOnly(t *testing.T) {
	e := setupWithNoChain(t)

	chain, err := e.Clu.DeployDefaultChain()
	require.NoError(t, err)

	chEnv := newChainEnv(t, e.Clu, chain)

	tx := chEnv.deployNativeIncCounterSC()

	// test calling root.FuncFindContractByName view function using client
	ret, err := chain.Cluster.WaspClient(0).CallView(
		chain.ChainID, root.Contract.Hname(), root.ViewFindContract.Name,
		dict.Dict{
			root.ParamHname: isc.Hn(nativeIncCounterSCName).Bytes(),
		})
	require.NoError(t, err)
	recb, err := ret.Get(root.ParamContractRecData)
	require.NoError(t, err)
	rec, err := root.ContractRecordFromBytes(recb)
	require.NoError(t, err)
	require.EqualValues(t, "testing contract deployment with inccounter", rec.Description)

	{
		txID, err := tx.ID()
		require.NoError(t, err)
		rec, err := chain.GetRequestReceipt(isc.NewRequestID(txID, 0))
		require.NoError(t, err)
		require.Nil(t, rec.Error)
	}
}

func TestDeployContractAndSpawn(t *testing.T) {
	e := setupWithNoChain(t)

	chain, err := e.Clu.DeployDefaultChain()
	require.NoError(t, err)

	chEnv := newChainEnv(t, e.Clu, chain)

	chEnv.deployNativeIncCounterSC()

	hname := isc.Hn(nativeIncCounterSCName)

	nameNew := "spawnedContract"
	dscrNew := "spawned contract it is"
	hnameNew := isc.Hn(nameNew)
	// send 'spawn' request to the SC which was just deployed
	par := chainclient.NewPostRequestParams(
		inccounter.VarName, nameNew,
		inccounter.VarDescription, dscrNew,
	).WithBaseTokens(100)
	tx, err := chain.OriginatorClient().Post1Request(hname, inccounter.FuncSpawn.Hname(), *par)
	require.NoError(t, err)

	receipts, err := chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(chain.ChainID, tx, 30*time.Second)
	require.NoError(t, err)
	require.Len(t, receipts, 1)

	chEnv.checkCoreContracts()
	for _, i := range chain.CommitteeNodes {
		blockIndex, err := chain.BlockIndex(i)
		require.NoError(t, err)
		require.Greater(t, blockIndex, uint32(2))

		contractRegistry, err := chain.ContractRegistry(i)
		require.NoError(t, err)
		require.EqualValues(t, len(corecontracts.All)+2, len(contractRegistry))

		cr := contractRegistry[hnameNew]
		require.EqualValues(t, dscrNew, cr.Description)
		require.EqualValues(t, nameNew, cr.Name)

		counterValue, err := chain.GetCounterValue(hname, i)
		require.NoError(t, err)
		require.EqualValues(t, 42, counterValue)
	}
}
