package tests

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes/cbalances"
	"testing"
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/tools/cluster"
	"github.com/stretchr/testify/require"
)

func TestBasicAccounts(t *testing.T) {
	setup(t, "test_cluster")
	counter, err := clu.StartMessageCounter(map[string]int{
		"chainrec":            2,
		"active_committee":    1,
		"dismissed_committee": 0,
		"state":               2,
		"request_in":          1,
		"request_out":         2,
	})
	check(err, t)
	defer counter.Close()
	chain, err := clu.DeployDefaultChain()
	check(err, t)
	testBasicAccounts(t, chain, counter)
}

func TestBasicAccountsN1(t *testing.T) {
	setup(t, "test_cluster")
	chainNodes := []int{0}
	counter, err := cluster.NewMessageCounter(clu, chainNodes, map[string]int{
		"chainrec":            2,
		"active_committee":    1,
		"dismissed_committee": 0,
		"state":               2,
		"request_in":          1,
		"request_out":         2,
	})
	check(err, t)
	defer counter.Close()
	chain, err := clu.DeployChain("single_node_chain", chainNodes, 1)
	check(err, t)
	testBasicAccounts(t, chain, counter)
}

func testBasicAccounts(t *testing.T, chain *cluster.Chain, counter *cluster.MessageCounter) {
	name := "inncounter1"
	hname := coretypes.Hn(name)
	description := "testing contract deployment with inccounter"
	programHash := inccounter.Interface.ProgramHash

	_, err = chain.DeployContract(name, programHash.String(), description, map[string]interface{}{
		inccounter.VarCounter: 42,
		root.ParamName:        name,
	})
	check(err, t)

	if !counter.WaitUntilExpectationsMet() {
		t.Fail()
	}

	t.Logf("   %s: %s", root.Name, root.Interface.Hname().String())
	t.Logf("   %s: %s", accounts.Name, accounts.Interface.Hname().String())

	chain.WithSCState(root.Interface.Hname(), func(host string, blockIndex uint32, state dict.Dict) bool {
		require.EqualValues(t, 2, blockIndex)
		checkRoots(t, chain)

		contractRegistry := collections.NewMapReadOnly(state, root.VarContractRegistry)
		require.EqualValues(t, 5, contractRegistry.MustLen())

		crBytes := contractRegistry.MustGetAt(hname.Bytes())
		require.NotNil(t, crBytes)
		cr, err := root.DecodeContractRecord(crBytes)
		check(err, t)

		require.EqualValues(t, programHash, cr.ProgramHash)
		require.EqualValues(t, description, cr.Description)
		require.EqualValues(t, 0, cr.OwnerFee)
		require.EqualValues(t, name, cr.Name)

		return true
	})

	chain.WithSCState(hname, func(host string, blockIndex uint32, state dict.Dict) bool {
		counterValue, _, _ := codec.DecodeInt64(state.MustGet(inccounter.VarCounter))
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
	chClient := chainclient.New(clu.Level1Client(), clu.WaspClient(0), chain.ChainID, scOwner.SigScheme())
	reqTx, err := chClient.PostRequest(hname, coretypes.Hn(inccounter.FuncIncCounter), chainclient.PostRequestParams{
		Transfer: cbalances.NewIotasOnly(transferIotas),
	})
	check(err, t)

	err = chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(reqTx, 30*time.Second)
	check(err, t)

	chain.WithSCState(hname, func(host string, blockIndex uint32, state dict.Dict) bool {
		counterValue, _, _ := codec.DecodeInt64(state.MustGet(inccounter.VarCounter))
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

func TestBasic2Accounts(t *testing.T) {
	setup(t, "test_cluster")

	counter, err := clu.StartMessageCounter(map[string]int{
		"chainrec":            2,
		"active_committee":    1,
		"dismissed_committee": 0,
		"state":               2,
		"request_in":          1,
		"request_out":         2,
	})
	check(err, t)
	defer counter.Close()

	chain, err := clu.DeployDefaultChain()
	check(err, t)

	name := "inncounter1"
	hname := coretypes.Hn(name)
	description := "testing contract deployment with inccounter"
	programHash := inccounter.Interface.ProgramHash
	check(err, t)

	_, err = chain.DeployContract(name, programHash.String(), description, map[string]interface{}{
		inccounter.VarCounter: 42,
		root.ParamName:        name,
	})
	check(err, t)

	if !counter.WaitUntilExpectationsMet() {
		t.Fail()
	}

	t.Logf("   %s: %s", root.Name, root.Interface.Hname().String())
	t.Logf("   %s: %s", accounts.Name, accounts.Interface.Hname().String())

	chain.WithSCState(root.Interface.Hname(), func(host string, blockIndex uint32, state dict.Dict) bool {
		require.EqualValues(t, 2, blockIndex)
		checkRoots(t, chain)

		contractRegistry := collections.NewMapReadOnly(state, root.VarContractRegistry)
		require.EqualValues(t, 5, contractRegistry.MustLen())

		crBytes := contractRegistry.MustGetAt(hname.Bytes())
		require.NotNil(t, crBytes)
		cr, err := root.DecodeContractRecord(crBytes)
		check(err, t)

		require.EqualValues(t, programHash, cr.ProgramHash)
		require.EqualValues(t, description, cr.Description)
		require.EqualValues(t, 0, cr.OwnerFee)
		require.EqualValues(t, name, cr.Name)

		return true
	})

	chain.WithSCState(hname, func(host string, blockIndex uint32, state dict.Dict) bool {
		counterValue, _, _ := codec.DecodeInt64(state.MustGet(inccounter.VarCounter))
		require.EqualValues(t, 42, counterValue)
		return true
	})

	if !clu.VerifyAddressBalances(&chain.Address, 3, map[balance.Color]int64{
		balance.ColorIOTA: 2,
		chain.Color:       1,
	}, "chain after deployment") {
		t.Fail()
	}

	originatorSigScheme := chain.OriginatorSigScheme()
	originatorAddress := chain.OriginatorAddress()

	if !clu.VerifyAddressBalances(originatorAddress, testutil.RequestFundsAmount-3, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 3, // 1 for chain, 1 init, 1 inccounter
	}, "originator after deployment") {
		t.Fail()
	}
	checkLedger(t, chain)

	myWallet := wallet.WithIndex(3)
	myWalletAddr := myWallet.Address()

	err = requestFunds(clu, myWalletAddr, "myWalletAddress")
	check(err, t)

	transferIotas := int64(42)
	myWalletClient := chainclient.New(clu.Level1Client(), clu.WaspClient(0), chain.ChainID, myWallet.SigScheme())
	reqTx, err := myWalletClient.PostRequest(hname, coretypes.Hn(inccounter.FuncIncCounter), chainclient.PostRequestParams{
		Transfer: cbalances.NewIotasOnly(transferIotas),
	})
	check(err, t)

	err = chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(reqTx, 30*time.Second)
	check(err, t)
	checkLedger(t, chain)

	chain.WithSCState(hname, func(host string, blockIndex uint32, state dict.Dict) bool {
		counterValue, _, _ := codec.DecodeInt64(state.MustGet(inccounter.VarCounter))
		require.EqualValues(t, 43, counterValue)
		return true
	})
	if !clu.VerifyAddressBalances(originatorAddress, testutil.RequestFundsAmount-3, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 3, // 1 for chain, 1 init, 1 inccounter
	}, "originator after") {
		t.Fail()
	}
	if !clu.VerifyAddressBalances(myWalletAddr, testutil.RequestFundsAmount-1-transferIotas, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 1 - transferIotas,
	}, "myWalletAddr after") {
		t.Fail()
	}
	if !clu.VerifyAddressBalances(&chain.Address, 4+transferIotas, map[balance.Color]int64{
		balance.ColorIOTA: 3 + transferIotas,
		chain.Color:       1,
	}, "chain after") {
		t.Fail()
	}
	// verify and print chain accounts
	s := "\n"
	agentID := coretypes.NewAgentIDFromContractID(coretypes.NewContractID(chain.ChainID, hname))
	s += fmt.Sprintf("contract: %s\n", agentID.String())
	actual := getAgentBalanceOnChain(t, chain, agentID, balance.ColorIOTA)
	require.EqualValues(t, 42, actual)

	agentID = coretypes.NewAgentIDFromAddress(*myWalletAddr)
	s += fmt.Sprintf("scOwner: %s\n", agentID.String())
	actual = getAgentBalanceOnChain(t, chain, agentID, balance.ColorIOTA)
	require.EqualValues(t, 1, actual) // 1 request sent, 1 iota from request

	agentID = coretypes.NewAgentIDFromAddress(*originatorAddress)
	s += fmt.Sprintf("originator: %s\n\n", agentID.String())
	actual = getAgentBalanceOnChain(t, chain, agentID, balance.ColorIOTA)
	require.EqualValues(t, 2, actual) // 1 request + 1 chain

	printAccounts(t, chain, "withdraw before")

	// withdraw back 2 iotas to originator address
	fmt.Printf("\norig addres from sigsheme: %s\n", originatorSigScheme.Address().String())
	originatorClient := chainclient.New(clu.Level1Client(), clu.WaspClient(0), chain.ChainID, originatorSigScheme)
	reqTx2, err := originatorClient.PostRequest(accounts.Interface.Hname(), coretypes.Hn(accounts.FuncWithdrawToAddress))
	check(err, t)

	err = chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(reqTx2, 30*time.Second)
	check(err, t)

	checkLedger(t, chain)

	printAccounts(t, chain, "withdraw after")

	// must remain 0 on chain
	agentID = coretypes.NewAgentIDFromAddress(*originatorAddress)
	actual = getAgentBalanceOnChain(t, chain, agentID, balance.ColorIOTA)
	require.EqualValues(t, 0, actual)

	if !clu.VerifyAddressBalances(originatorAddress, testutil.RequestFundsAmount-1, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 1,
	}, "originator after withdraw: "+originatorAddress.String()) {
		t.Fail()
	}
}
