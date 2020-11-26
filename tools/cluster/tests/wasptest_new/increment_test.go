package wasptest

import (
	"bytes"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
	"github.com/iotaledger/wasp/packages/vm/examples/inccounter"
	"github.com/iotaledger/wasp/plugins/wasmtimevm"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

const incName = "increment"
const incDescription = "Increment, a PoC smart contract"

var incHname = coretypes.Hn(incName)

func TestIncDeployment(t *testing.T) {
	clu, chain := setupAndLoad(t, incName, incDescription, 0, nil)

	if !clu.WaitUntilExpectationsMet() {
		t.Fail()
	}

	chain.WithSCState(root.Hname, func(host string, blockIndex uint32, state codec.ImmutableMustCodec) bool {
		require.EqualValues(t, 2, blockIndex)

		chid, _ := state.GetChainID(root.VarChainID)
		require.EqualValues(t, &chain.ChainID, chid)

		aid, _ := state.GetAgentID(root.VarChainOwnerID)
		require.EqualValues(t, *chain.OriginatorID(), *aid)

		desc, _ := state.GetString(root.VarDescription)
		require.EqualValues(t, chain.Description, desc)

		contractRegistry := state.GetMap(root.VarContractRegistry)
		require.EqualValues(t, 3, contractRegistry.Len())
		//--
		crBytes := contractRegistry.GetAt(root.Hname.Bytes())
		require.NotNil(t, crBytes)
		require.True(t, bytes.Equal(crBytes, util.MustBytes(&root.RootContractRecord)))
		//--
		crBytes = contractRegistry.GetAt(incHname.Bytes())
		require.NotNil(t, crBytes)
		cr, err := root.DecodeContractRecord(crBytes)
		check(err, t)

		require.EqualValues(t, wasmtimevm.PluginName, cr.VMType)
		require.EqualValues(t, incName, cr.Name)
		require.EqualValues(t, incDescription, cr.Description)
		require.EqualValues(t, 0, cr.NodeFee)
		return true
	})
	chain.WithSCState(incHname, func(host string, blockIndex uint32, state codec.ImmutableMustCodec) bool {
		counterValue, _ := state.GetInt64(inccounter.VarCounter)
		require.EqualValues(t, 0, counterValue)
		return true
	})
}

func TestIncNothing(t *testing.T) {
	testNothing(t, 1)
}

func TestInc5xNothing(t *testing.T) {
	testNothing(t, 5)
}

func testNothing(t *testing.T, numRequests int) {
	clu, chain := setupAndLoad(t, incName, incDescription, numRequests, nil)

	entryPoint := coretypes.Hn("nothing")
	for i := 0; i < numRequests; i++ {
		tx, err := chain.OriginatorClient().PostRequest(incHname, entryPoint, nil, nil, nil)
		check(err, t)
		err = chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(tx, 30*time.Second)
		check(err, t)
	}

	if !clu.WaitUntilExpectationsMet() {
		t.Fail()
	}

	chain.WithSCState(root.Hname, func(host string, blockIndex uint32, state codec.ImmutableMustCodec) bool {
		require.EqualValues(t, numRequests+2, blockIndex)

		chid, _ := state.GetChainID(root.VarChainID)
		require.EqualValues(t, &chain.ChainID, chid)

		aid, _ := state.GetAgentID(root.VarChainOwnerID)
		require.EqualValues(t, *chain.OriginatorID(), *aid)

		desc, _ := state.GetString(root.VarDescription)
		require.EqualValues(t, chain.Description, desc)

		contractRegistry := state.GetMap(root.VarContractRegistry)
		require.EqualValues(t, 3, contractRegistry.Len())
		//--
		crBytes := contractRegistry.GetAt(root.Hname.Bytes())
		require.NotNil(t, crBytes)
		require.True(t, bytes.Equal(crBytes, util.MustBytes(&root.RootContractRecord)))
		//--
		crBytes = contractRegistry.GetAt(incHname.Bytes())
		require.NotNil(t, crBytes)
		cr, err := root.DecodeContractRecord(crBytes)
		check(err, t)
		require.EqualValues(t, wasmtimevm.PluginName, cr.VMType)
		require.EqualValues(t, incName, cr.Name)
		require.EqualValues(t, incDescription, cr.Description)
		require.EqualValues(t, 0, cr.NodeFee)
		return true
	})
	chain.WithSCState(incHname, func(host string, blockIndex uint32, state codec.ImmutableMustCodec) bool {
		counterValue, _ := state.GetInt64(inccounter.VarCounter)
		require.EqualValues(t, 0, counterValue)
		return true
	})
}

func TestIncIncrement(t *testing.T) {
	testIncrement(t, 1)
}

func TestInc5xIncrement(t *testing.T) {
	testIncrement(t, 5)
}

func testIncrement(t *testing.T, numRequests int) {
	clu, chain := setupAndLoad(t, incName, incDescription, numRequests, nil)

	entryPoint := coretypes.Hn("increment")
	for i := 0; i < numRequests; i++ {
		tx, err := chain.OriginatorClient().PostRequest(incHname, entryPoint, nil, nil, nil)
		check(err, t)
		err = chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(tx, 30*time.Second)
		check(err, t)
	}

	if !clu.WaitUntilExpectationsMet() {
		t.Fail()
	}

	chain.WithSCState(root.Hname, func(host string, blockIndex uint32, state codec.ImmutableMustCodec) bool {
		require.EqualValues(t, numRequests+2, blockIndex)

		chid, _ := state.GetChainID(root.VarChainID)
		require.EqualValues(t, &chain.ChainID, chid)

		aid, _ := state.GetAgentID(root.VarChainOwnerID)
		require.EqualValues(t, *chain.OriginatorID(), *aid)

		desc, _ := state.GetString(root.VarDescription)
		require.EqualValues(t, chain.Description, desc)

		contractRegistry := state.GetMap(root.VarContractRegistry)
		require.EqualValues(t, 3, contractRegistry.Len())
		//--
		crBytes := contractRegistry.GetAt(root.Hname.Bytes())
		require.NotNil(t, crBytes)
		require.True(t, bytes.Equal(crBytes, util.MustBytes(&root.RootContractRecord)))
		//--
		crBytes = contractRegistry.GetAt(incHname.Bytes())
		require.NotNil(t, crBytes)
		cr, err := root.DecodeContractRecord(crBytes)
		check(err, t)
		require.EqualValues(t, wasmtimevm.PluginName, cr.VMType)
		require.EqualValues(t, incName, cr.Name)
		require.EqualValues(t, incDescription, cr.Description)
		require.EqualValues(t, 0, cr.NodeFee)
		return true
	})
	chain.WithSCState(incHname, func(host string, blockIndex uint32, state codec.ImmutableMustCodec) bool {
		counterValue, _ := state.GetInt64(inccounter.VarCounter)
		require.EqualValues(t, numRequests, counterValue)
		return true
	})
}

func TestIncrementWithTransfer(t *testing.T) {
	clu, chain := setupAndLoad(t, incName, incDescription, 1, nil)

	if !clu.VerifyAddressBalances(&chain.Address, 3, map[balance.Color]int64{
		balance.ColorIOTA: 2,
		chain.Color:       1,
	}, "chain after deployment") {
		t.Fail()
	}

	err := requestFunds(clu, scOwnerAddr, "originator")
	check(err, t)

	chClient := chainclient.New(clu.NodeClient, clu.WaspClient(0), chain.ChainID, scOwner.SigScheme())

	entryPoint := coretypes.Hn("increment")
	transfer := map[balance.Color]int64{
		balance.ColorIOTA: 42,
	}
	tx, err := chClient.PostRequest(incHname, entryPoint, nil, transfer, nil)
	check(err, t)

	err = chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(tx, 30*time.Second)
	check(err, t)

	if !clu.WaitUntilExpectationsMet() {
		t.Fail()
	}

	if !clu.VerifyAddressBalances(scOwnerAddr, testutil.RequestFundsAmount-1-42, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 1 - 42,
	}, "owner after") {
		t.Fail()
	}
	if !clu.VerifyAddressBalances(&chain.Address, 4+42, map[balance.Color]int64{
		balance.ColorIOTA: 3 + 42,
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
	require.EqualValues(t, 2, actual) // 1 request sent

	chain.WithSCState(incHname, func(host string, blockIndex uint32, state codec.ImmutableMustCodec) bool {
		counterValue, _ := state.GetInt64(inccounter.VarCounter)
		require.EqualValues(t, 1, counterValue)
		return true
	})
}

func TestIncCallIncrement(t *testing.T) {
	clu, chain := setupAndLoad(t, incName, incDescription, 1, nil)

	err := requestFunds(clu, scOwnerAddr, "originator")
	check(err, t)

	entryPoint := coretypes.Hn("incrementCallIncrement")
	tx, err := chain.OriginatorClient().PostRequest(incHname, entryPoint, nil, nil, nil)
	check(err, t)
	err = chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(tx, 30*time.Second)
	check(err, t)

	if !clu.WaitUntilExpectationsMet() {
		t.Fail()
	}

	chain.WithSCState(incHname, func(host string, blockIndex uint32, state codec.ImmutableMustCodec) bool {
		counterValue, _ := state.GetInt64(inccounter.VarCounter)
		require.EqualValues(t, 2, counterValue)
		return true
	})
}

func TestIncPostIncrement(t *testing.T) {
	clu, chain := setupAndLoad(t, incName, incDescription, 2, nil)

	err := requestFunds(clu, scOwnerAddr, "originator")
	check(err, t)

	chClient := chainclient.New(clu.NodeClient, clu.WaspClient(0), chain.ChainID, scOwner.SigScheme())

	entryPoint := coretypes.Hn("incrementPostIncrement")
	transfer := map[balance.Color]int64{
		balance.ColorIOTA: 42,
	}
	// here we post request 'incrementPostIncrement' to the inccounter SC which then in turn will post 'increment' request to itself
	// so it needs one iota to create the tx but the log shows the following error:
	// ERROR   consensus/action.go:188 checking result tx properties: sc transaction must contain exactly one chain token output
	tx, err := chClient.PostRequest(incHname, entryPoint, nil, transfer, nil)
	check(err, t)

	err = chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(tx, 30*time.Second)
	check(err, t)

	if !clu.WaitUntilExpectationsMet() {
		t.Fail()
	}

	chain.WithSCState(incHname, func(host string, blockIndex uint32, state codec.ImmutableMustCodec) bool {
		counterValue, _ := state.GetInt64(inccounter.VarCounter)
		require.EqualValues(t, 2, counterValue)
		return true
	})
}

func TestIncRepeatManyIncrement(t *testing.T) {
	const numRepeats = 5
	clu, chain := setupAndLoad(t, incName, incDescription, 2, nil)

	//TODO transfer 5i to contract
	entryPoint := coretypes.Hn("incrementRepeatMany")
	tx, err := chain.OriginatorClient().PostRequest(incHname, entryPoint, nil, nil, map[string]interface{}{
		inccounter.VarNumRepeats: numRepeats,
	})

	check(err, t)
	err = chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(tx, 30*time.Second)
	check(err, t)

	if !clu.WaitUntilExpectationsMet() {
		t.Fail()
	}

	chain.WithSCState(incHname, func(host string, blockIndex uint32, state codec.ImmutableMustCodec) bool {
		counterValue, _ := state.GetInt64(inccounter.VarCounter)
		require.EqualValues(t, numRepeats+1, counterValue)
		repeats, _ := state.GetInt64(inccounter.VarNumRepeats)
		require.EqualValues(t, 0, repeats)
		return true
	})
}
