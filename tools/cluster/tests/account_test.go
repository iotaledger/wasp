package tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/tools/cluster"
	"github.com/stretchr/testify/require"
)

func TestBasicAccounts(t *testing.T) {
	setup(t, "test_cluster")
	counter1, err := clu.StartMessageCounter(map[string]int{
		"state":       2,
		"request_in":  0,
		"request_out": 1,
	})
	check(err, t)
	defer counter1.Close()
	chain1, err := clu.DeployDefaultChain()
	check(err, t)
	testBasicAccounts(t, chain1, counter1)
}

func TestBasicAccountsN1(t *testing.T) {
	setup(t, "test_cluster")
	chainNodes := []int{0}
	counter1, err := cluster.NewMessageCounter(clu, chainNodes, map[string]int{
		"state": 3,
	})
	check(err, t)
	defer counter1.Close()
	chain1, err := clu.DeployChainWithDKG("single_node_chain", chainNodes, chainNodes, 1)
	check(err, t)
	testBasicAccounts(t, chain1, counter1)
}

func testBasicAccounts(t *testing.T, chain *cluster.Chain, counter *cluster.MessageCounter) {
	hname := iscp.Hn(incCounterSCName)
	description := "testing contract deployment with inccounter"
	programHash1 := inccounter.Contract.ProgramHash

	_, err = chain.DeployContract(incCounterSCName, programHash1.String(), description, map[string]interface{}{
		inccounter.VarCounter: 42,
		root.ParamName:        incCounterSCName,
	})
	check(err, t)

	if !counter.WaitUntilExpectationsMet() {
		t.Fail()
	}

	t.Logf("   %s: %s", root.Contract.Name, root.Contract.Hname().String())
	t.Logf("   %s: %s", accounts.Contract.Name, accounts.Contract.Hname().String())

	checkCoreContracts(t, chain)

	for i := range chain.CommitteeNodes {
		blockIndex, err := chain.BlockIndex(i)
		require.NoError(t, err)
		require.EqualValues(t, 2, blockIndex)

		contractRegistry, err := chain.ContractRegistry(i)
		require.NoError(t, err)

		cr := contractRegistry[hname]

		require.EqualValues(t, programHash1, cr.ProgramHash)
		require.EqualValues(t, description, cr.Description)
		require.EqualValues(t, 0, cr.OwnerFee)
		require.EqualValues(t, incCounterSCName, cr.Name)

		counterValue, err := chain.GetCounterValue(hname, i)
		require.NoError(t, err)
		require.EqualValues(t, 42, counterValue)
	}

	if !clu.VerifyAddressBalances(chain.ChainID.AsAddress(), ledgerstate.DustThresholdAliasOutputIOTA+2, map[ledgerstate.Color]uint64{
		ledgerstate.ColorIOTA: ledgerstate.DustThresholdAliasOutputIOTA + 2,
	}, "chain after deployment") {
		t.Fail()
	}

	err = requestFunds(clu, scOwnerAddr, "originator")
	check(err, t)

	transferIotas := uint64(42)
	chClient := chainclient.New(clu.GoshimmerClient(), clu.WaspClient(0), chain.ChainID, scOwner)

	par := chainclient.NewPostRequestParams().WithIotas(transferIotas)
	reqTx, err := chClient.Post1Request(hname, inccounter.FuncIncCounter.Hname(), *par)
	check(err, t)

	err = chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(chain.ChainID, reqTx, 10*time.Second)
	check(err, t)

	for i := range chain.CommitteeNodes {
		counterValue, err := chain.GetCounterValue(incCounterSCHname, i)
		require.NoError(t, err)
		require.EqualValues(t, 43, counterValue)
	}

	if !clu.VerifyAddressBalances(scOwnerAddr, solo.Saldo-transferIotas, map[ledgerstate.Color]uint64{
		ledgerstate.ColorIOTA: solo.Saldo - transferIotas,
	}, "owner after") {
		t.Fail()
	}

	if !clu.VerifyAddressBalances(chain.ChainID.AsAddress(), ledgerstate.DustThresholdAliasOutputIOTA+transferIotas+2, map[ledgerstate.Color]uint64{
		ledgerstate.ColorIOTA: ledgerstate.DustThresholdAliasOutputIOTA + transferIotas + 2,
	}, "chain after") {
		t.Fail()
	}
	agentID := iscp.NewAgentID(chain.ChainID.AsAddress(), hname)
	actual := getBalanceOnChain(t, chain, agentID, ledgerstate.ColorIOTA)
	require.EqualValues(t, 42, actual)
}

func TestBasic2Accounts(t *testing.T) {
	setup(t, "test_cluster")

	counter1, err := clu.StartMessageCounter(map[string]int{
		"state":       2,
		"request_in":  0,
		"request_out": 1,
	})
	check(err, t)
	defer counter1.Close()

	chain1, err := clu.DeployDefaultChain()
	check(err, t)

	hname := iscp.Hn(incCounterSCName)
	description := "testing contract deployment with inccounter"
	programHash1 := inccounter.Contract.ProgramHash
	check(err, t)

	_, err = chain1.DeployContract(incCounterSCName, programHash1.String(), description, map[string]interface{}{
		inccounter.VarCounter: 42,
		root.ParamName:        incCounterSCName,
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

		t.Logf("%+v", contractRegistry)
		cr := contractRegistry[hname]

		require.EqualValues(t, programHash1, cr.ProgramHash)
		require.EqualValues(t, description, cr.Description)
		require.EqualValues(t, 0, cr.OwnerFee)
		require.EqualValues(t, incCounterSCName, cr.Name)

		counterValue, err := chain1.GetCounterValue(hname, i)
		require.NoError(t, err)
		require.EqualValues(t, 42, counterValue)
	}

	if !clu.VerifyAddressBalances(chain1.ChainID.AsAddress(), ledgerstate.DustThresholdAliasOutputIOTA+2, map[ledgerstate.Color]uint64{
		ledgerstate.ColorIOTA: ledgerstate.DustThresholdAliasOutputIOTA + 2,
	}, "chain after deployment") {
		t.Fail()
	}

	originatorSigScheme := chain1.OriginatorKeyPair()
	originatorAddress := chain1.OriginatorAddress()

	if !clu.VerifyAddressBalances(originatorAddress, solo.Saldo-ledgerstate.DustThresholdAliasOutputIOTA-2, map[ledgerstate.Color]uint64{
		ledgerstate.ColorIOTA: solo.Saldo - ledgerstate.DustThresholdAliasOutputIOTA - 2, //
	}, "originator after deployment") {
		t.Fail()
	}
	checkLedger(t, chain1)

	myWallet := wallet.KeyPair(3)
	myWalletAddr := ledgerstate.NewED25519Address(myWallet.PublicKey)

	err = requestFunds(clu, myWalletAddr, "myWalletAddress")
	check(err, t)

	transferIotas := uint64(42)
	myWalletClient := chainclient.New(clu.GoshimmerClient(), clu.WaspClient(0), chain1.ChainID, myWallet)

	par := chainclient.NewPostRequestParams().WithIotas(transferIotas)
	reqTx, err := myWalletClient.Post1Request(hname, inccounter.FuncIncCounter.Hname(), *par)
	check(err, t)

	err = chain1.CommitteeMultiClient().WaitUntilAllRequestsProcessed(chain1.ChainID, reqTx, 30*time.Second)
	check(err, t)
	checkLedger(t, chain1)

	for _, i := range chain1.CommitteeNodes {
		counterValue, err := chain1.GetCounterValue(hname, i)
		require.NoError(t, err)
		require.EqualValues(t, 43, counterValue)
	}
	if !clu.VerifyAddressBalances(originatorAddress, solo.Saldo-ledgerstate.DustThresholdAliasOutputIOTA-2, map[ledgerstate.Color]uint64{
		ledgerstate.ColorIOTA: solo.Saldo - ledgerstate.DustThresholdAliasOutputIOTA - 2, // 1 for chain, 1 init, 1 inccounter
	}, "originator after") {
		t.Fail()
	}
	if !clu.VerifyAddressBalances(myWalletAddr, solo.Saldo-transferIotas, map[ledgerstate.Color]uint64{
		ledgerstate.ColorIOTA: solo.Saldo - transferIotas,
	}, "myWalletAddr after") {
		t.Fail()
	}
	if !clu.VerifyAddressBalances(chain1.ChainID.AsAddress(), ledgerstate.DustThresholdAliasOutputIOTA+2+transferIotas, map[ledgerstate.Color]uint64{
		ledgerstate.ColorIOTA: ledgerstate.DustThresholdAliasOutputIOTA + 2 + transferIotas,
	}, "chain after") {
		t.Fail()
	}
	// verify and print chain accounts
	agentID := iscp.NewAgentID(chain1.ChainID.AsAddress(), hname)
	actual := getBalanceOnChain(t, chain1, agentID, ledgerstate.ColorIOTA)
	require.EqualValues(t, 42, actual)

	printAccounts(t, chain1, "withdraw before")

	// withdraw back 2 iotas to originator address
	fmt.Printf("\norig address from sigsheme: %s\n", originatorAddress.Base58())
	originatorClient := chainclient.New(clu.GoshimmerClient(), clu.WaspClient(0), chain1.ChainID, originatorSigScheme)
	reqTx2, err := originatorClient.Post1Request(accounts.Contract.Hname(), accounts.FuncWithdraw.Hname())
	check(err, t)

	err = chain1.CommitteeMultiClient().WaitUntilAllRequestsProcessed(chain1.ChainID, reqTx2, 30*time.Second)
	check(err, t)

	checkLedger(t, chain1)

	printAccounts(t, chain1, "withdraw after")

	// must remain 0 on chain
	agentID = iscp.NewAgentID(originatorAddress, 0)
	actual = getBalanceOnChain(t, chain1, agentID, ledgerstate.ColorIOTA)
	require.EqualValues(t, 0, actual)

	if !clu.VerifyAddressBalances(originatorAddress, solo.Saldo-ledgerstate.DustThresholdAliasOutputIOTA-3, map[ledgerstate.Color]uint64{
		ledgerstate.ColorIOTA: solo.Saldo - ledgerstate.DustThresholdAliasOutputIOTA - 3,
	}, "originator after withdraw: "+originatorAddress.String()) {
		t.Fail()
	}
}
