package tests

import (
	"testing"
	"time"

	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/stretchr/testify/require"
)

func TestDeployChain(t *testing.T) {
	setup(t, "test_cluster")

	counter1, err := clu.StartMessageCounter(map[string]int{
		"dismissed_chain": 0,
		"state":           2,
		"request_out":     1,
	})
	check(err, t)
	defer counter1.Close()

	chain1, err := clu.DeployDefaultChain()
	check(err, t)

	if !counter1.WaitUntilExpectationsMet() {
		t.Fail()
	}
	chainID, chainOwnerID := getChainInfo(t, chain1)
	require.Equal(t, chainID, chain1.ChainID)
	require.Equal(t, chainOwnerID, *coretypes.NewAgentID(chain1.OriginatorAddress(), 0))
	t.Logf("--- chainID: %s", chainID.String())
	t.Logf("--- chainOwnerID: %s", chainOwnerID.String())

	checkCoreContracts(t, chain1)
	checkRootsOutside(t, chain1)
	for _, i := range chain1.CommitteeNodes {
		blockIndex, err := chain1.BlockIndex(i)
		require.NoError(t, err)
		require.EqualValues(t, 1, blockIndex)

		contractRegistry, err := chain1.ContractRegistry(i)
		require.NoError(t, err)
		require.EqualValues(t, len(core.AllCoreContractsByHash), len(contractRegistry))
	}
}

func TestDeployContractFail(t *testing.T) {
	setup(t, "test_cluster")

	counter1, err := clu.StartMessageCounter(map[string]int{
		"dismissed_chain": 0,
		"state":           2,
		"request_out":     1,
	})
	check(err, t)
	defer counter1.Close()

	chain1, err := clu.DeployDefaultChain()
	check(err, t)

	tx, err := chain1.DeployContract(incCounterSCName, programHash.String(), "", map[string]interface{}{
		root.ParamName: "", // intentionally empty so that it fails
	})
	check(err, t)

	if !counter1.WaitUntilExpectationsMet() {
		t.Fail()
	}

	checkCoreContracts(t, chain1)
	for _, i := range chain1.CommitteeNodes {
		blockIndex, err := chain1.BlockIndex(i)
		require.NoError(t, err)
		require.EqualValues(t, 2, blockIndex)

		contractRegistry, err := chain1.ContractRegistry(i)
		require.NoError(t, err)
		// contract was not deployed:
		require.EqualValues(t, len(core.AllCoreContractsByHash), len(contractRegistry))
	}

	// query error message from blocklog:
	rec, _, _, err := chain1.GetRequestLogRecord(coretypes.NewRequestID(tx.ID(), 0))
	require.NoError(t, err)
	require.Contains(t, string(rec.LogData), "wrong name")
}

func TestDeployContractOnly(t *testing.T) {
	setup(t, "test_cluster")

	counter1, err := clu.StartMessageCounter(map[string]int{
		"dismissed_committee": 0,
		"state":               2,
		"request_out":         1,
	})
	check(err, t)
	defer counter1.Close()

	chain1, err := clu.DeployDefaultChain()
	check(err, t)

	tx := deployIncCounterSC(t, chain1, counter1)

	// test calling root.FuncFindContractByName view function using client
	ret, err := chain1.Cluster.WaspClient(0).CallView(
		chain1.ChainID, root.Interface.Hname(), root.FuncFindContract,
		dict.Dict{
			root.ParamHname: coretypes.Hn(incCounterSCName).Bytes(),
		})
	check(err, t)
	recb, err := ret.Get(root.VarData)
	check(err, t)
	rec, err := root.DecodeContractRecord(recb)
	check(err, t)
	require.EqualValues(t, "testing contract deployment with inccounter", rec.Description)

	{
		rec, _, _, err := chain1.GetRequestLogRecord(coretypes.NewRequestID(tx.ID(), 0))
		require.NoError(t, err)
		require.Empty(t, string(rec.LogData))
	}
}

func TestDeployContractAndSpawn(t *testing.T) {
	setup(t, "test_cluster")

	counter1, err := clu.StartMessageCounter(map[string]int{
		"dismissed_committee": 0,
		"state":               2,
		"request_out":         1,
	})
	check(err, t)
	defer counter1.Close()

	chain1, err := clu.DeployDefaultChain()
	check(err, t)

	deployIncCounterSC(t, chain1, counter1)

	hname := coretypes.Hn(incCounterSCName)

	nameNew := "spawnedContract"
	dscrNew := "spawned contract it is"
	hnameNew := coretypes.Hn(nameNew)
	// send 'spawn' request to the SC which was just deployed
	par := chainclient.NewPostRequestParams(
		inccounter.VarName, nameNew,
		inccounter.VarDescription, dscrNew,
	).WithIotas(1)
	tx, err := chain1.OriginatorClient().Post1Request(hname, coretypes.Hn(inccounter.FuncSpawn), *par)
	check(err, t)

	err = chain1.CommitteeMultiClient().WaitUntilAllRequestsProcessed(chain1.ChainID, tx, 30*time.Second)
	check(err, t)

	checkCoreContracts(t, chain1)
	for _, i := range chain1.CommitteeNodes {
		blockIndex, err := chain1.BlockIndex(i)
		require.NoError(t, err)
		require.EqualValues(t, 3, blockIndex)

		contractRegistry, err := chain1.ContractRegistry(i)
		require.NoError(t, err)
		require.EqualValues(t, len(core.AllCoreContractsByHash)+2, len(contractRegistry))

		cr := contractRegistry[hnameNew]
		require.EqualValues(t, dscrNew, cr.Description)
		require.EqualValues(t, 0, cr.OwnerFee)
		require.EqualValues(t, nameNew, cr.Name)

		counterValue, err := chain1.GetCounterValue(hname, i)
		require.NoError(t, err)
		require.EqualValues(t, 42, counterValue)
	}
}
