package wasptest

import (
	"bytes"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
	"github.com/iotaledger/wasp/packages/vm/examples/inccounter"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

const incName = "inccounter"
const incDescription = "IncCounter, a PoC smart contract"

var incHname = coretypes.Hn(incName)

func checkCounter(t *testing.T, expected int) bool {
	return chain.WithSCState(incHname, func(host string, blockIndex uint32, state codec.ImmutableMustCodec) bool {
		counterValue, _ := state.GetInt64(inccounter.VarCounter)
		require.EqualValues(t, expected, counterValue)
		return true
	})
}

func TestIncDeployment(t *testing.T) {
	setupAndLoad(t, incName, incDescription, 0, nil)

	if !clu.WaitUntilExpectationsMet() {
		t.Fail()
	}

	chain.WithSCState(root.Interface.Hname(), func(host string, blockIndex uint32, state codec.ImmutableMustCodec) bool {
		require.EqualValues(t, 3, blockIndex)

		chid, _ := state.GetChainID(root.VarChainID)
		require.EqualValues(t, &chain.ChainID, chid)

		aid, _ := state.GetAgentID(root.VarChainOwnerID)
		require.EqualValues(t, *chain.OriginatorID(), *aid)

		desc, _ := state.GetString(root.VarDescription)
		require.EqualValues(t, chain.Description, desc)

		contractRegistry := state.GetMap(root.VarContractRegistry)
		require.EqualValues(t, 4, contractRegistry.Len())
		//--
		crBytes := contractRegistry.GetAt(root.Interface.Hname().Bytes())
		require.NotNil(t, crBytes)
		require.True(t, bytes.Equal(crBytes, util.MustBytes(&root.RootContractRecord)))
		//--
		crBytes = contractRegistry.GetAt(incHname.Bytes())
		require.NotNil(t, crBytes)
		cr, err := root.DecodeContractRecord(crBytes)
		check(err, t)

		require.EqualValues(t, programHash, cr.ProgramHash)
		require.EqualValues(t, incName, cr.Name)
		require.EqualValues(t, incDescription, cr.Description)
		require.EqualValues(t, 0, cr.NodeFee)
		return true
	})
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

	entryPoint := coretypes.Hn("nothing")
	for i := 0; i < numRequests; i++ {
		tx, err := client.PostRequest(incHname, entryPoint)
		check(err, t)
		err = chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(tx, 30*time.Second)
		check(err, t)
	}

	if !clu.WaitUntilExpectationsMet() {
		t.Fail()
	}

	chain.WithSCState(root.Interface.Hname(), func(host string, blockIndex uint32, state codec.ImmutableMustCodec) bool {
		require.EqualValues(t, numRequests+3, blockIndex)

		chid, _ := state.GetChainID(root.VarChainID)
		require.EqualValues(t, &chain.ChainID, chid)

		aid, _ := state.GetAgentID(root.VarChainOwnerID)
		require.EqualValues(t, *chain.OriginatorID(), *aid)

		desc, _ := state.GetString(root.VarDescription)
		require.EqualValues(t, chain.Description, desc)

		contractRegistry := state.GetMap(root.VarContractRegistry)
		require.EqualValues(t, 4, contractRegistry.Len())
		//--
		crBytes := contractRegistry.GetAt(root.Interface.Hname().Bytes())
		require.NotNil(t, crBytes)
		require.True(t, bytes.Equal(crBytes, util.MustBytes(&root.RootContractRecord)))
		//--
		crBytes = contractRegistry.GetAt(incHname.Bytes())
		require.NotNil(t, crBytes)
		cr, err := root.DecodeContractRecord(crBytes)
		check(err, t)
		require.EqualValues(t, programHash, cr.ProgramHash)
		require.EqualValues(t, incName, cr.Name)
		require.EqualValues(t, incDescription, cr.Description)
		require.EqualValues(t, 0, cr.NodeFee)
		return true
	})
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

	entryPoint := coretypes.Hn("increment")
	for i := 0; i < numRequests; i++ {
		tx, err := client.PostRequest(incHname, entryPoint)
		check(err, t)
		err = chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(tx, 30*time.Second)
		check(err, t)
	}

	if !clu.WaitUntilExpectationsMet() {
		t.Fail()
	}

	chain.WithSCState(root.Interface.Hname(), func(host string, blockIndex uint32, state codec.ImmutableMustCodec) bool {
		require.EqualValues(t, numRequests+3, blockIndex)

		chid, _ := state.GetChainID(root.VarChainID)
		require.EqualValues(t, &chain.ChainID, chid)

		aid, _ := state.GetAgentID(root.VarChainOwnerID)
		require.EqualValues(t, *chain.OriginatorID(), *aid)

		desc, _ := state.GetString(root.VarDescription)
		require.EqualValues(t, chain.Description, desc)

		contractRegistry := state.GetMap(root.VarContractRegistry)
		require.EqualValues(t, 4, contractRegistry.Len())
		//--
		crBytes := contractRegistry.GetAt(root.Interface.Hname().Bytes())
		require.NotNil(t, crBytes)
		require.True(t, bytes.Equal(crBytes, util.MustBytes(&root.RootContractRecord)))
		//--
		crBytes = contractRegistry.GetAt(incHname.Bytes())
		require.NotNil(t, crBytes)
		cr, err := root.DecodeContractRecord(crBytes)
		check(err, t)
		require.EqualValues(t, programHash, cr.ProgramHash)
		require.EqualValues(t, incName, cr.Name)
		require.EqualValues(t, incDescription, cr.Description)
		require.EqualValues(t, 0, cr.NodeFee)
		return true
	})
	checkCounter(t, numRequests)
}

func TestIncrementWithTransfer(t *testing.T) {
	setupAndLoad(t, incName, incDescription, 1, nil)

	if !clu.VerifyAddressBalances(&chain.Address, 4, map[balance.Color]int64{
		balance.ColorIOTA: 3,
		chain.Color:       1,
	}, "chain after deployment") {
		t.Fail()
	}

	entryPoint := coretypes.Hn("increment")
	postRequest(t, incHname, entryPoint, 42, nil)

	if !clu.VerifyAddressBalances(scOwnerAddr, testutil.RequestFundsAmount-1-42, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 1 - 42,
	}, "owner after") {
		t.Fail()
	}
	if !clu.VerifyAddressBalances(&chain.Address, 5+42, map[balance.Color]int64{
		balance.ColorIOTA: 4 + 42,
		chain.Color:       1,
	}, "chain after") {
		t.Fail()
	}
	agentID := coretypes.NewAgentIDFromContractID(coretypes.NewContractID(chain.ChainID, incHname))
	actual := getAgentBalanceOnChain(t, chain, agentID, balance.ColorIOTA)
	require.EqualValues(t, 42, actual)

	agentID = coretypes.NewAgentIDFromAddress(*scOwnerAddr)
	actual = getAgentBalanceOnChain(t, chain, agentID, balance.ColorIOTA)
	require.EqualValues(t, 1, actual) // 1 request sent

	agentID = coretypes.NewAgentIDFromAddress(*chain.OriginatorAddress())
	actual = getAgentBalanceOnChain(t, chain, agentID, balance.ColorIOTA)
	require.EqualValues(t, 3, actual) // 2 requests sent

	checkCounter(t, 1)
}

func TestIncCallIncrement1(t *testing.T) {
	setupAndLoad(t, incName, incDescription, 1, nil)

	entryPoint := coretypes.Hn("incrementCallIncrement")
	postRequest(t, incHname, entryPoint, 0, nil)

	checkCounter(t, 2)
}

func TestIncCallIncrement2Recurse5x(t *testing.T) {
	setupAndLoad(t, incName, incDescription, 1, nil)

	entryPoint := coretypes.Hn("incrementCallIncrementRecurse5x")
	postRequest(t, incHname, entryPoint, 0, nil)

	checkCounter(t, 6)
}

func TestIncPostIncrement(t *testing.T) {
	setupAndLoad(t, incName, incDescription, 3, nil)

	entryPoint := coretypes.Hn("incrementPostIncrement")
	postRequest(t, incHname, entryPoint, 1, nil)

	checkCounter(t, 2)
}

func TestIncRepeatManyIncrement(t *testing.T) {
	const numRepeats = 5
	setupAndLoad(t, incName, incDescription, numRepeats+2, nil)

	entryPoint := coretypes.Hn("incrementRepeatMany")
	postRequest(t, incHname, entryPoint, numRepeats, map[string]interface{}{
		inccounter.VarNumRepeats: numRepeats,
	})

	chain.WithSCState(incHname, func(host string, blockIndex uint32, state codec.ImmutableMustCodec) bool {
		counterValue, _ := state.GetInt64(inccounter.VarCounter)
		require.EqualValues(t, numRepeats+1, counterValue)
		repeats, _ := state.GetInt64(inccounter.VarNumRepeats)
		require.EqualValues(t, 0, repeats)
		return true
	})
}

func TestIncLocalStateInternalCall(t *testing.T) {
	setupAndLoad(t, incName, incDescription, 1, nil)
	entryPoint := coretypes.Hn("incrementLocalStateInternalCall")
	postRequest(t, incHname, entryPoint, 0, nil)
	checkCounter(t, 2)
}

func TestIncLocalStateSandboxCall(t *testing.T) {
	setupAndLoad(t, incName, incDescription, 1, nil)
	entryPoint := coretypes.Hn("incrementLocalStateSandboxCall")
	postRequest(t, incHname, entryPoint, 0, nil)
	checkCounter(t, 0)
}

func TestIncLocalStatePost(t *testing.T) {
	setupAndLoad(t, incName, incDescription, 5, nil)
	entryPoint := coretypes.Hn("incrementLocalStatePost")
	postRequest(t, incHname, entryPoint, 3, nil)
	checkCounter(t, 0)
}

func TestIncViewCounter(t *testing.T) {
	setupAndLoad(t, incName, incDescription, 1, nil)
	entryPoint := coretypes.Hn("increment")
	postRequest(t, incHname, entryPoint, 0, nil)
	checkCounter(t, 1)
	ret, err := chain.Cluster.WaspClient(0).CallView(
		chain.ContractID(coretypes.Hn(incName)),
		"incrementViewCounter",
		nil,
	)
	check(err, t)

	c := codec.NewCodec(ret)
	counter, _, err := c.GetInt64("counter")
	check(err, t)
	require.EqualValues(t, 1, counter)
}
