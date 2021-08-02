package tests

import (
	"testing"
	"time"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/colored"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/core/root/rootimpl"
	"github.com/stretchr/testify/require"
)

func (e *contractEnv) checkSC(numRequests int) {
	for i := range e.chain.CommitteeNodes {
		blockIndex, err := e.chain.BlockIndex(i)
		require.NoError(e.t, err)
		require.EqualValues(e.t, numRequests+3, blockIndex)

		cl := e.chain.SCClient(root.Contract.Hname(), nil, i)
		ret, err := cl.CallView(root.FuncGetChainInfo.Name, nil)
		require.NoError(e.t, err)

		chid, _, _ := codec.DecodeChainID(ret.MustGet(root.VarChainID))
		require.EqualValues(e.t, e.chain.ChainID, chid)

		aid, _, _ := codec.DecodeAgentID(ret.MustGet(root.VarChainOwnerID))
		require.EqualValues(e.t, *e.chain.OriginatorID(), aid)

		desc, _, _ := codec.DecodeString(ret.MustGet(root.VarDescription))
		require.EqualValues(e.t, e.chain.Description, desc)

		contractRegistry, err := rootimpl.DecodeContractRegistry(collections.NewMapReadOnly(ret, root.VarContractRegistry))
		require.NoError(e.t, err)
		require.EqualValues(e.t, len(core.AllCoreContractsByHash)+1, len(contractRegistry))

		cr := contractRegistry[incHname]
		require.EqualValues(e.t, e.programHash, cr.ProgramHash)
		require.EqualValues(e.t, incName, cr.Name)
		require.EqualValues(e.t, incDescription, cr.Description)
		require.EqualValues(e.t, 0, cr.OwnerFee)
	}
}

func (e *chainEnv) checkCounter(expected int) {
	for i := range e.chain.CommitteeNodes {
		counterValue, err := e.chain.GetCounterValue(incHname, i)
		require.NoError(e.t, err)
		require.EqualValues(e.t, expected, counterValue)
	}
}

func TestIncDeployment(t *testing.T) {
	e := setupWithContractAndMessageCounter(t, incName, incDescription, 0)

	if !e.counter.WaitUntilExpectationsMet() {
		t.Fail()
	}

	for i := range e.chain.CommitteeNodes {
		blockIndex, err := e.chain.BlockIndex(i)
		require.NoError(t, err)
		require.EqualValues(t, 3, blockIndex)

		cl := e.chain.SCClient(root.Contract.Hname(), nil, i)
		ret, err := cl.CallView(root.FuncGetChainInfo.Name, nil)
		require.NoError(t, err)

		chid, _, _ := codec.DecodeChainID(ret.MustGet(root.VarChainID))
		require.EqualValues(t, e.chain.ChainID, chid)

		aid, _, _ := codec.DecodeAgentID(ret.MustGet(root.VarChainOwnerID))
		require.EqualValues(t, *e.chain.OriginatorID(), aid)

		desc, _, _ := codec.DecodeString(ret.MustGet(root.VarDescription))
		require.EqualValues(t, e.chain.Description, desc)

		contractRegistry, err := rootimpl.DecodeContractRegistry(collections.NewMapReadOnly(ret, root.VarContractRegistry))
		require.NoError(t, err)
		require.EqualValues(t, len(core.AllCoreContractsByHash)+1, len(contractRegistry))

		cr := contractRegistry[incHname]

		require.EqualValues(t, e.programHash, cr.ProgramHash)
		require.EqualValues(t, incName, cr.Name)
		require.EqualValues(t, incDescription, cr.Description)
		require.EqualValues(t, 0, cr.OwnerFee)
	}
	e.checkCounter(0)
}

func TestIncNothing(t *testing.T) {
	testNothing(t, 1)
}

func TestInc5xNothing(t *testing.T) {
	testNothing(t, 5)
}

func testNothing(t *testing.T, numRequests int) {
	e := setupWithContractAndMessageCounter(t, incName, incDescription, numRequests)

	entryPoint := iscp.Hn("nothing")
	for i := 0; i < numRequests; i++ {
		tx, err := e.chainClient().Post1Request(incHname, entryPoint)
		require.NoError(t, err)
		err = e.chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(e.chain.ChainID, tx, 30*time.Second)
		require.NoError(t, err)
	}

	if !e.counter.WaitUntilExpectationsMet() {
		t.Fail()
	}

	e.checkSC(numRequests)
	e.checkCounter(0)
}

func TestIncIncrement(t *testing.T) {
	testIncrement(t, 1)
}

func TestInc5xIncrement(t *testing.T) {
	testIncrement(t, 5)
}

func testIncrement(t *testing.T, numRequests int) {
	e := setupWithContractAndMessageCounter(t, incName, incDescription, numRequests)

	entryPoint := iscp.Hn("increment")
	for i := 0; i < numRequests; i++ {
		tx, err := e.chainClient().Post1Request(incHname, entryPoint)
		require.NoError(t, err)
		err = e.chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(e.chain.ChainID, tx, 30*time.Second)
		require.NoError(t, err)
	}

	if !e.counter.WaitUntilExpectationsMet() {
		t.Fail()
	}

	e.checkSC(numRequests)
	e.checkCounter(numRequests)
}

func TestIncrementWithTransfer(t *testing.T) {
	e := setupWithContractAndMessageCounter(t, incName, incDescription, 1)

	entryPoint := iscp.Hn("increment")
	e.postRequest(incHname, entryPoint, 42, nil)

	if !e.clu.VerifyAddressBalances(scOwnerAddr, solo.Saldo-42,
		colored.NewBalancesForIotas(solo.Saldo-42),
		"owner after") {
		t.Fail()
	}
	agentID := iscp.NewAgentID(e.chain.ChainID.AsAddress(), incHname)
	actual := e.getBalanceOnChain(agentID, colored.IOTA)
	require.EqualValues(t, 42, actual)

	agentID = iscp.NewAgentID(scOwnerAddr, 0)
	actual = e.getBalanceOnChain(agentID, colored.IOTA)
	require.EqualValues(t, 0, actual)

	e.checkCounter(1)
}

func TestIncCallIncrement1(t *testing.T) {
	e := setupWithContractAndMessageCounter(t, incName, incDescription, 1)

	entryPoint := iscp.Hn("callIncrement")
	e.postRequest(incHname, entryPoint, 1, nil)

	e.checkCounter(2)
}

func TestIncCallIncrement2Recurse5x(t *testing.T) {
	e := setupWithContractAndMessageCounter(t, incName, incDescription, 1)

	entryPoint := iscp.Hn("callIncrementRecurse5x")
	e.postRequest(incHname, entryPoint, 0, nil)

	e.checkCounter(6)
}

func TestIncPostIncrement(t *testing.T) {
	e := setupWithContractAndMessageCounter(t, incName, incDescription, 3)

	entryPoint := iscp.Hn("postIncrement")
	e.postRequest(incHname, entryPoint, 1, nil)

	e.checkCounter(2)
}

func TestIncRepeatManyIncrement(t *testing.T) {
	const numRepeats = 5
	e := setupWithContractAndMessageCounter(t, incName, incDescription, numRepeats+2)

	entryPoint := iscp.Hn("repeatMany")
	e.postRequest(incHname, entryPoint, numRepeats, map[string]interface{}{
		varNumRepeats: numRepeats,
	})

	for i := range e.chain.CommitteeNodes {
		b, err := e.chain.GetStateVariable(incHname, varCounter, i)
		require.NoError(t, err)
		counterValue, _, _ := codec.DecodeInt64(b)
		require.EqualValues(t, numRepeats+1, counterValue)

		b, err = e.chain.GetStateVariable(incHname, varNumRepeats, i)
		require.NoError(t, err)
		repeats, _, _ := codec.DecodeInt64(b)
		require.EqualValues(t, 0, repeats)
	}
}

func TestIncLocalStateInternalCall(t *testing.T) {
	e := setupWithContractAndMessageCounter(t, incName, incDescription, 1)
	entryPoint := iscp.Hn("localStateInternalCall")
	e.postRequest(incHname, entryPoint, 0, nil)
	e.checkCounter(2)
}

func TestIncLocalStateSandboxCall(t *testing.T) {
	e := setupWithContractAndMessageCounter(t, incName, incDescription, 1)
	entryPoint := iscp.Hn("localStateSandboxCall")
	e.postRequest(incHname, entryPoint, 0, nil)
	e.checkCounter(0)
}

func TestIncLocalStatePost(t *testing.T) {
	e := setupWithContractAndMessageCounter(t, incName, incDescription, 3)
	entryPoint := iscp.Hn("localStatePost")
	e.postRequest(incHname, entryPoint, 3, nil)
	e.checkCounter(0)
}

func TestIncViewCounter(t *testing.T) {
	e := setupWithContractAndMessageCounter(t, incName, incDescription, 1)
	entryPoint := iscp.Hn("increment")
	e.postRequest(incHname, entryPoint, 0, nil)
	e.checkCounter(1)
	ret, err := e.chain.Cluster.WaspClient(0).CallView(
		e.chain.ChainID, incHname, "getCounter", nil,
	)
	require.NoError(t, err)

	counter, _, err := codec.DecodeInt64(ret.MustGet(varCounter))
	require.NoError(t, err)
	require.EqualValues(t, 1, counter)
}
