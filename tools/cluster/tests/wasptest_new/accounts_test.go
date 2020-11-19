package wasptest

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/accountsc"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
	"github.com/iotaledger/wasp/packages/vm/examples"
	"github.com/iotaledger/wasp/packages/vm/examples/inccounter"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestBasicAccounts(t *testing.T) {
	clu := setup(t, "test_cluster")

	err := clu.ListenToMessages(map[string]int{
		"chainrec":            2,
		"active_committee":    1,
		"dismissed_committee": 0,
		"state":               2,
		"request_in":          1,
		"request_out":         2,
	})
	check(err, t)

	chain, err := clu.DeployDefaultChain()
	check(err, t)

	name := "inncounter1"
	hname := coretypes.Hn(name)
	description := "testing contract deployment with inccounter"

	_, err = chain.DeployBuiltinContract(name, examples.VMType, inccounter.ProgramHashStr, description, map[string]interface{}{
		inccounter.VarCounter: 42,
		root.ParamName:        name,
	})
	check(err, t)

	if !clu.WaitUntilExpectationsMet() {
		t.Fail()
	}

	t.Logf("   %s: %s", root.ContractName, root.Hname.String())
	t.Logf("   %s: %s", accountsc.ContractName, accountsc.Hname.String())

	chain.WithSCState(root.Hname, func(host string, blockIndex uint32, state codec.ImmutableMustCodec) bool {
		require.EqualValues(t, 2, blockIndex)
		checkRoots(t, chain)

		contractRegistry := state.GetMap(root.VarContractRegistry)
		require.EqualValues(t, 3, contractRegistry.Len())

		crBytes := contractRegistry.GetAt(hname.Bytes())
		require.NotNil(t, crBytes)
		cr, err := root.DecodeContractRecord(crBytes)
		check(err, t)

		require.EqualValues(t, examples.VMType, cr.VMType)
		require.EqualValues(t, description, cr.Description)
		require.EqualValues(t, 0, cr.NodeFee)
		require.EqualValues(t, name, cr.Name)

		return true
	})

	chain.WithSCState(hname, func(host string, blockIndex uint32, state codec.ImmutableMustCodec) bool {
		counterValue, _ := state.GetInt64(inccounter.VarCounter)
		require.EqualValues(t, 42, counterValue)
		return true
	})

	if !clu.VerifyAddressBalances(&chain.Address, 3, map[balance.Color]int64{
		balance.ColorIOTA: 2,
		chain.Color:       1,
	}, "chain after deployment") {
		t.Fail()
	}

	err = requestFunds(clu, scOwnerAddr, "originator")
	check(err, t)

	transferIotas := int64(42)
	chClient := chainclient.New(clu.NodeClient, clu.WaspClient(0), &chain.ChainID, scOwner.SigScheme())
	reqTx, err := chClient.PostRequest(hname, inccounter.EntryPointIncCounter, nil, map[balance.Color]int64{
		balance.ColorIOTA: transferIotas,
	}, nil)
	check(err, t)

	err = chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(reqTx, 30*time.Second)
	check(err, t)

	chain.WithSCState(hname, func(host string, blockIndex uint32, state codec.ImmutableMustCodec) bool {
		counterValue, _ := state.GetInt64(inccounter.VarCounter)
		require.EqualValues(t, 43, counterValue)
		return true
	})
	if !clu.VerifyAddressBalances(scOwnerAddr, testutil.RequestFundsAmount-1-transferIotas, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 1 - transferIotas,
	}, "owner after") {
		t.Fail()
	}

	if !clu.VerifyAddressBalances(&chain.Address, 4+transferIotas, map[balance.Color]int64{
		balance.ColorIOTA: 3 + transferIotas,
		chain.Color:       1,
	}, "chain after") {
		t.Fail()
	}
	agentID := coretypes.NewAgentIDFromContractID(coretypes.NewContractID(chain.ChainID, hname))
	actual := getAgentBalanceOnChain(t, chain, agentID, balance.ColorIOTA)
	require.EqualValues(t, 42, actual)

	agentID = coretypes.NewAgentIDFromAddress(*scOwnerAddr)
	actual = getAgentBalanceOnChain(t, chain, agentID, balance.ColorIOTA)
	require.EqualValues(t, 1, actual) // 1 request sent

	agentID = coretypes.NewAgentIDFromAddress(*chain.OriginatorAddress())
	actual = getAgentBalanceOnChain(t, chain, agentID, balance.ColorIOTA)
	require.EqualValues(t, 2, actual) // 1 request sent
}
