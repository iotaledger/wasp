package tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/solo"
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

	if !clu.VerifyAddressBalances(chain.Address, 3, map[ledgerstate.Color]uint64{
		ledgerstate.ColorIOTA: 2,
	}, "chain after deployment") {
		t.Fail()
	}

	err = requestFunds(clu, scOwnerAddr, "originator")
	check(err, t)

	transferIotas := uint64(42)
	chClient := chainclient.New(clu.GoshimmerClient(), clu.WaspClient(0), chain.ChainID, scOwner)
	reqTx, err := chClient.PostRequest(hname, coretypes.Hn(inccounter.FuncIncCounter), chainclient.PostRequestParams{
		Transfer: coretypes.NewTransferIotas(transferIotas),
	})
	check(err, t)

	err = chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(chain.ChainID, reqTx, 30*time.Second)
	check(err, t)

	chain.WithSCState(hname, func(host string, blockIndex uint32, state dict.Dict) bool {
		counterValue, _, _ := codec.DecodeInt64(state.MustGet(inccounter.VarCounter))
		require.EqualValues(t, 43, counterValue)
		return true
	})
	if !clu.VerifyAddressBalances(scOwnerAddr, solo.Saldo-1-transferIotas, map[ledgerstate.Color]uint64{
		ledgerstate.ColorIOTA: solo.Saldo - 1 - transferIotas,
	}, "owner after") {
		t.Fail()
	}

	if !clu.VerifyAddressBalances(chain.Address, 4+transferIotas, map[ledgerstate.Color]uint64{
		ledgerstate.ColorIOTA: 3 + transferIotas,
	}, "chain after") {
		t.Fail()
	}
	agentID := coretypes.NewAgentID(chain.ChainID.AsAddress(), hname)
	actual := getAgentBalanceOnChain(t, chain, agentID, ledgerstate.ColorIOTA)
	require.EqualValues(t, 42, actual)

	agentID = coretypes.NewAgentID(scOwnerAddr, 0)
	actual = getAgentBalanceOnChain(t, chain, agentID, ledgerstate.ColorIOTA)
	require.EqualValues(t, 1, actual) // 1 request sent

	agentID = coretypes.NewAgentID(chain.OriginatorAddress(), 0)
	actual = getAgentBalanceOnChain(t, chain, agentID, ledgerstate.ColorIOTA)
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

	if !clu.VerifyAddressBalances(chain.Address, 3, map[ledgerstate.Color]uint64{
		ledgerstate.ColorIOTA: 2,
	}, "chain after deployment") {
		t.Fail()
	}

	originatorSigScheme := chain.OriginatorKeyPair()
	originatorAddress := chain.OriginatorAddress()

	if !clu.VerifyAddressBalances(originatorAddress, solo.Saldo-3, map[ledgerstate.Color]uint64{
		ledgerstate.ColorIOTA: solo.Saldo - 3, // 1 for chain, 1 init, 1 inccounter
	}, "originator after deployment") {
		t.Fail()
	}
	checkLedger(t, chain)

	myWallet := wallet.KeyPair(3)
	myWalletAddr := ledgerstate.NewED25519Address(myWallet.PublicKey)

	err = requestFunds(clu, myWalletAddr, "myWalletAddress")
	check(err, t)

	transferIotas := uint64(42)
	myWalletClient := chainclient.New(clu.GoshimmerClient(), clu.WaspClient(0), chain.ChainID, myWallet)
	reqTx, err := myWalletClient.PostRequest(hname, coretypes.Hn(inccounter.FuncIncCounter), chainclient.PostRequestParams{
		Transfer: coretypes.NewTransferIotas(transferIotas),
	})
	check(err, t)

	err = chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(chain.ChainID, reqTx, 30*time.Second)
	check(err, t)
	checkLedger(t, chain)

	chain.WithSCState(hname, func(host string, blockIndex uint32, state dict.Dict) bool {
		counterValue, _, _ := codec.DecodeInt64(state.MustGet(inccounter.VarCounter))
		require.EqualValues(t, 43, counterValue)
		return true
	})
	if !clu.VerifyAddressBalances(originatorAddress, solo.Saldo-3, map[ledgerstate.Color]uint64{
		ledgerstate.ColorIOTA: solo.Saldo - 3, // 1 for chain, 1 init, 1 inccounter
	}, "originator after") {
		t.Fail()
	}
	if !clu.VerifyAddressBalances(myWalletAddr, solo.Saldo-1-transferIotas, map[ledgerstate.Color]uint64{
		ledgerstate.ColorIOTA: solo.Saldo - 1 - transferIotas,
	}, "myWalletAddr after") {
		t.Fail()
	}
	if !clu.VerifyAddressBalances(chain.Address, 4+transferIotas, map[ledgerstate.Color]uint64{
		ledgerstate.ColorIOTA: 3 + transferIotas,
	}, "chain after") {
		t.Fail()
	}
	// verify and print chain accounts
	s := "\n"
	agentID := coretypes.NewAgentID(chain.ChainID.AsAddress(), hname)
	s += fmt.Sprintf("contract: %s\n", agentID.String())
	actual := getAgentBalanceOnChain(t, chain, agentID, ledgerstate.ColorIOTA)
	require.EqualValues(t, 42, actual)

	agentID = coretypes.NewAgentID(myWalletAddr, 0)
	s += fmt.Sprintf("scOwner: %s\n", agentID.String())
	actual = getAgentBalanceOnChain(t, chain, agentID, ledgerstate.ColorIOTA)
	require.EqualValues(t, 1, actual) // 1 request sent, 1 iota from request

	agentID = coretypes.NewAgentID(originatorAddress, 0)
	s += fmt.Sprintf("originator: %s\n\n", agentID.String())
	actual = getAgentBalanceOnChain(t, chain, agentID, ledgerstate.ColorIOTA)
	require.EqualValues(t, 2, actual) // 1 request + 1 chain

	printAccounts(t, chain, "withdraw before")

	// withdraw back 2 iotas to originator address
	fmt.Printf("\norig addres from sigsheme: %s\n", originatorAddress.Base58())
	originatorClient := chainclient.New(clu.GoshimmerClient(), clu.WaspClient(0), chain.ChainID, originatorSigScheme)
	reqTx2, err := originatorClient.PostRequest(accounts.Interface.Hname(), coretypes.Hn(accounts.FuncWithdraw))
	check(err, t)

	err = chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(chain.ChainID, reqTx2, 30*time.Second)
	check(err, t)

	checkLedger(t, chain)

	printAccounts(t, chain, "withdraw after")

	// must remain 0 on chain
	agentID = coretypes.NewAgentID(originatorAddress, 0)
	actual = getAgentBalanceOnChain(t, chain, agentID, ledgerstate.ColorIOTA)
	require.EqualValues(t, 0, actual)

	if !clu.VerifyAddressBalances(originatorAddress, solo.Saldo-1, map[ledgerstate.Color]uint64{
		ledgerstate.ColorIOTA: solo.Saldo - 1,
	}, "originator after withdraw: "+originatorAddress.String()) {
		t.Fail()
	}
}
