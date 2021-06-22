package tests

import (
	"testing"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/solo"

	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
)

func TestDepositWithdraw(t *testing.T) {
	setup(t, "test_cluster")

	chain, err := clu.DeployDefaultChain()
	check(err, t)

	testOwner := wallet.KeyPair(1)
	myAddress := ledgerstate.NewED25519Address(testOwner.PublicKey)

	err = requestFunds(clu, myAddress, "myAddress")
	check(err, t)
	if !clu.VerifyAddressBalances(myAddress, solo.Saldo, map[ledgerstate.Color]uint64{
		ledgerstate.ColorIOTA: solo.Saldo,
	}, "myAddress begin") {
		t.Fail()
	}
	if !clu.VerifyAddressBalances(chain.OriginatorAddress(), solo.Saldo-ledgerstate.DustThresholdAliasOutputIOTA-1, map[ledgerstate.Color]uint64{
		ledgerstate.ColorIOTA: solo.Saldo - ledgerstate.DustThresholdAliasOutputIOTA - 1,
	}, "originatorAddress begin") {
		t.Fail()
	}
	if !clu.VerifyAddressBalances(chain.ChainAddress(), ledgerstate.DustThresholdAliasOutputIOTA+1, map[ledgerstate.Color]uint64{
		ledgerstate.ColorIOTA: ledgerstate.DustThresholdAliasOutputIOTA + 1,
	}, "chainAddress begin") {
		t.Fail()
	}
	checkLedger(t, chain)

	myAgentID := coretypes.NewAgentID(myAddress, 0)
	origAgentId := coretypes.NewAgentID(chain.OriginatorAddress(), 0)

	checkBalanceOnChain(t, chain, origAgentId, ledgerstate.ColorIOTA, 0)
	checkBalanceOnChain(t, chain, myAgentID, ledgerstate.ColorIOTA, 0)
	checkLedger(t, chain)

	// deposit some iotas to the chain
	depositIotas := uint64(42)
	chClient := chainclient.New(clu.GoshimmerClient(), clu.WaspClient(0), chain.ChainID, testOwner)

	par := chainclient.NewPostRequestParams().WithIotas(depositIotas)
	reqTx, err := chClient.Post1Request(accounts.Interface.Hname(), coretypes.Hn(accounts.FuncDeposit), *par)
	check(err, t)

	err = chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(chain.ChainID, reqTx, 30*time.Second)
	check(err, t)
	checkLedger(t, chain)
	checkBalanceOnChain(t, chain, myAgentID, ledgerstate.ColorIOTA, depositIotas)
	checkBalanceOnChain(t, chain, origAgentId, ledgerstate.ColorIOTA, 0)

	if !clu.VerifyAddressBalances(myAddress, solo.Saldo-depositIotas, map[ledgerstate.Color]uint64{
		ledgerstate.ColorIOTA: solo.Saldo - depositIotas,
	}, "myAddress after deposit") {
		t.Fail()
	}

	// withdraw iotas back
	reqTx3, err := chClient.Post1Request(accounts.Interface.Hname(), coretypes.Hn(accounts.FuncWithdraw))
	check(err, t)
	err = chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(chain.ChainID, reqTx3, 30*time.Second)
	check(err, t)

	check(err, t)
	checkLedger(t, chain)
	checkBalanceOnChain(t, chain, myAgentID, ledgerstate.ColorIOTA, 0)

	if !clu.VerifyAddressBalances(myAddress, solo.Saldo-1, map[ledgerstate.Color]uint64{
		ledgerstate.ColorIOTA: solo.Saldo - 1,
	}, "myAddress after withdraw") {
		t.Fail()
	}
}
