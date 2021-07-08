package tests

import (
	"testing"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/tools/cluster"
	"github.com/stretchr/testify/require"
)

func checkSC(t *testing.T, chain *cluster.Chain, numRequests int) {
	for i := range chain.CommitteeNodes {
		blockIndex, err := chain.BlockIndex(i)
		require.NoError(t, err)
		require.EqualValues(t, numRequests+3, blockIndex)

		cl := chain.SCClient(root.Interface.Hname(), nil, i)
		ret, err := cl.CallView(root.FuncGetChainInfo)
		require.NoError(t, err)

		chid, _, _ := codec.DecodeChainID(ret.MustGet(root.VarChainID))
		require.EqualValues(t, chain.ChainID, chid)

		aid, _, _ := codec.DecodeAgentID(ret.MustGet(root.VarChainOwnerID))
		require.EqualValues(t, *chain.OriginatorID(), aid)

		desc, _, _ := codec.DecodeString(ret.MustGet(root.VarDescription))
		require.EqualValues(t, chain.Description, desc)

		contractRegistry, err := root.DecodeContractRegistry(collections.NewMapReadOnly(ret, root.VarContractRegistry))
		require.NoError(t, err)
		require.EqualValues(t, len(core.AllCoreContractsByHash)+1, len(contractRegistry))

		cr := contractRegistry[incHname]
		require.EqualValues(t, programHash, cr.ProgramHash)
		require.EqualValues(t, incName, cr.Name)
		require.EqualValues(t, incDescription, cr.Description)
		require.EqualValues(t, 0, cr.OwnerFee)
	}
}

func checkCounter(t *testing.T, expected int) {
	for i := range chain.CommitteeNodes {
		counterValue, err := chain.GetCounterValue(incHname, i)
		require.NoError(t, err)
		require.EqualValues(t, expected, counterValue)
	}
}

func TestIncDeployment(t *testing.T) {
	setupAndLoad(t, incName, incDescription, 0, nil)
	defer counter.Close()

	if !counter.WaitUntilExpectationsMet() {
		t.Fail()
	}

	for i := range chain.CommitteeNodes {
		blockIndex, err := chain.BlockIndex(i)
		require.NoError(t, err)
		require.EqualValues(t, 3, blockIndex)

		cl := chain.SCClient(root.Interface.Hname(), nil, i)
		ret, err := cl.CallView(root.FuncGetChainInfo)
		require.NoError(t, err)

		chid, _, _ := codec.DecodeChainID(ret.MustGet(root.VarChainID))
		require.EqualValues(t, chain.ChainID, chid)

		aid, _, _ := codec.DecodeAgentID(ret.MustGet(root.VarChainOwnerID))
		require.EqualValues(t, *chain.OriginatorID(), aid)

		desc, _, _ := codec.DecodeString(ret.MustGet(root.VarDescription))
		require.EqualValues(t, chain.Description, desc)

		contractRegistry, err := root.DecodeContractRegistry(collections.NewMapReadOnly(ret, root.VarContractRegistry))
		require.NoError(t, err)
		require.EqualValues(t, len(core.AllCoreContractsByHash)+1, len(contractRegistry))

		cr := contractRegistry[incHname]

		require.EqualValues(t, programHash, cr.ProgramHash)
		require.EqualValues(t, incName, cr.Name)
		require.EqualValues(t, incDescription, cr.Description)
		require.EqualValues(t, 0, cr.OwnerFee)
	}
	checkCounter(t, 0)
}

func TestIncNothing(t *testing.T) {
	testNothing(t, 1)
}

func TestInc5xNothing(t *testing.T) {
	testNothing(t, 5)
}

func testNothing(t *testing.T, numRequests int) {
	setupAndLoad(t, incName, incDescription, numRequests, nil)
	defer counter.Close()

	entryPoint := coretypes.Hn("nothing")
	for i := 0; i < numRequests; i++ {
		tx, err := client.Post1Request(incHname, entryPoint)
		check(err, t)
		err = chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(chain.ChainID, tx, 30*time.Second)
		check(err, t)
	}

	if !counter.WaitUntilExpectationsMet() {
		t.Fail()
	}

	checkSC(t, chain, numRequests)
	checkCounter(t, 0)
}

func TestIncIncrement(t *testing.T) {
	testIncrement(t, 1)
}

func TestInc5xIncrement(t *testing.T) {
	testIncrement(t, 5)
}

func testIncrement(t *testing.T, numRequests int) {
	setupAndLoad(t, incName, incDescription, numRequests, nil)
	defer counter.Close()

	entryPoint := coretypes.Hn("increment")
	for i := 0; i < numRequests; i++ {
		tx, err := client.Post1Request(incHname, entryPoint)
		check(err, t)
		err = chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(chain.ChainID, tx, 30*time.Second)
		check(err, t)
	}

	if !counter.WaitUntilExpectationsMet() {
		t.Fail()
	}

	checkSC(t, chain, numRequests)
	checkCounter(t, numRequests)
}

func TestIncrementWithTransfer(t *testing.T) {
	setupAndLoad(t, incName, incDescription, 1, nil)

	entryPoint := coretypes.Hn("increment")
	postRequest(t, incHname, entryPoint, 42, nil)

	if !clu.VerifyAddressBalances(scOwnerAddr, solo.Saldo-42, map[ledgerstate.Color]uint64{
		ledgerstate.ColorIOTA: solo.Saldo - 42,
	}, "owner after") {
		t.Fail()
	}
	agentID := coretypes.NewAgentID(chain.ChainID.AsAddress(), incHname)
	actual := getBalanceOnChain(t, chain, agentID, ledgerstate.ColorIOTA)
	require.EqualValues(t, 42, actual)

	agentID = coretypes.NewAgentID(scOwnerAddr, 0)
	actual = getBalanceOnChain(t, chain, agentID, ledgerstate.ColorIOTA)
	require.EqualValues(t, 0, actual)

	checkCounter(t, 1)
}

func TestIncCallIncrement1(t *testing.T) {
	setupAndLoad(t, incName, incDescription, 1, nil)

	entryPoint := coretypes.Hn("call_increment")
	postRequest(t, incHname, entryPoint, 0, nil)

	checkCounter(t, 2)
}

func TestIncCallIncrement2Recurse5x(t *testing.T) {
	setupAndLoad(t, incName, incDescription, 1, nil)

	entryPoint := coretypes.Hn("call_increment_recurse5x")
	postRequest(t, incHname, entryPoint, 0, nil)

	checkCounter(t, 6)
}

func TestIncPostIncrement(t *testing.T) {
	setupAndLoad(t, incName, incDescription, 3, nil)

	entryPoint := coretypes.Hn("postIncrement")
	postRequest(t, incHname, entryPoint, 1, nil)

	checkCounter(t, 2)
}

func TestIncRepeatManyIncrement(t *testing.T) {
	const numRepeats = 5
	setupAndLoad(t, incName, incDescription, numRepeats+2, nil)

	entryPoint := coretypes.Hn("repeatMany")
	postRequest(t, incHname, entryPoint, numRepeats, map[string]interface{}{
		varNumRepeats: numRepeats,
	})

	for i := range chain.CommitteeNodes {
		b, err := chain.GetStateVariable(incHname, varCounter, i)
		require.NoError(t, err)
		counterValue, _, _ := codec.DecodeInt64(b)
		require.EqualValues(t, numRepeats+1, counterValue)

		b, err = chain.GetStateVariable(incHname, varNumRepeats, i)
		require.NoError(t, err)
		repeats, _, _ := codec.DecodeInt64(b)
		require.EqualValues(t, 0, repeats)
	}
}

func TestIncLocalStateInternalCall(t *testing.T) {
	setupAndLoad(t, incName, incDescription, 1, nil)
	entryPoint := coretypes.Hn("localStateInternalCall")
	postRequest(t, incHname, entryPoint, 0, nil)
	checkCounter(t, 2)
}

func TestIncLocalStateSandboxCall(t *testing.T) {
	setupAndLoad(t, incName, incDescription, 1, nil)
	entryPoint := coretypes.Hn("localStateSandboxCall")
	postRequest(t, incHname, entryPoint, 0, nil)
	checkCounter(t, 0)
}

func TestIncLocalStatePost(t *testing.T) {
	setupAndLoad(t, incName, incDescription, 5, nil)
	entryPoint := coretypes.Hn("localStatePost")
	postRequest(t, incHname, entryPoint, 3, nil)
	checkCounter(t, 0)
}

func TestIncViewCounter(t *testing.T) {
	setupAndLoad(t, incName, incDescription, 1, nil)
	entryPoint := coretypes.Hn("increment")
	postRequest(t, incHname, entryPoint, 0, nil)
	checkCounter(t, 1)
	ret, err := chain.Cluster.WaspClient(0).CallView(
		chain.ChainID, incHname, "getCounter",
	)
	check(err, t)

	counter, _, err := codec.DecodeInt64(ret.MustGet(varCounter))
	check(err, t)
	require.EqualValues(t, 1, counter)
}
