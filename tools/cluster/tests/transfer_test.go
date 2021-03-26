package tests

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/solo"
	"testing"
	"time"

	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
)

func TestDepositWithdraw(t *testing.T) {
	setup(t, "test_cluster")

	chain, err := clu.DeployDefaultChain()
	check(err, t)

	testOwner := wallet.WithIndex(1)
	mySigScheme := testOwner.SigScheme()
	myAddress := testOwner.Address()

	err = requestFunds(clu, myAddress, "myAddress")
	check(err, t)
	if !clu.VerifyAddressBalances(myAddress, solo.Saldo, map[ledgerstate.Color]uint64{
		ledgerstate.ColorIOTA: solo.Saldo,
	}, "myAddress begin") {
		t.Fail()
	}
	if !clu.VerifyAddressBalances(chain.OriginatorAddress(), solo.Saldo-2, map[ledgerstate.Color]uint64{
		ledgerstate.ColorIOTA: solo.Saldo - 2,
	}, "originatorAddress begin") {
		t.Fail()
	}
	if !clu.VerifyAddressBalances(chain.ChainAddress(), 2, map[ledgerstate.Color]uint64{
		chain.Color:           1,
		ledgerstate.ColorIOTA: 1,
	}, "chainAddress begin") {
		t.Fail()
	}
	checkLedger(t, chain)

	myAgentID := coretypes.NewAgentID(myAddress, 0)
	origAgentId := coretypes.NewAgentID(chain.OriginatorAddress(), 0)

	checkBalanceOnChain(t, chain, origAgentId, ledgerstate.ColorIOTA, 1)
	checkBalanceOnChain(t, chain, myAgentID, ledgerstate.ColorIOTA, 0)
	checkLedger(t, chain)

	// deposit some iotas to the chain
	depositIotas := uint64(42)
	chClient := chainclient.New(clu.Level1Client(), clu.WaspClient(0), chain.ChainID, mySigScheme)
	reqTx, err := chClient.PostRequest(accounts.Interface.Hname(), coretypes.Hn(accounts.FuncDeposit), chainclient.PostRequestParams{
		Transfer: coretypes.NewTransferIotas(depositIotas),
	})
	check(err, t)
	err = chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(reqTx, 30*time.Second)
	check(err, t)
	checkLedger(t, chain)
	checkBalanceOnChain(t, chain, myAgentID, ledgerstate.ColorIOTA, depositIotas+1) // 1 iota from request
	checkBalanceOnChain(t, chain, origAgentId, ledgerstate.ColorIOTA, 1)

	if !clu.VerifyAddressBalances(myAddress, solo.Saldo-depositIotas-1, map[ledgerstate.Color]uint64{
		ledgerstate.ColorIOTA: solo.Saldo - depositIotas - 1,
	}, "myAddress after deposit") {
		t.Fail()
	}

	// withdraw iotas back
	reqTx3, err := chClient.PostRequest(accounts.Interface.Hname(), coretypes.Hn(accounts.FuncWithdraw))
	check(err, t)
	err = chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(reqTx3, 30*time.Second)
	check(err, t)

	check(err, t)
	checkLedger(t, chain)
	checkBalanceOnChain(t, chain, myAgentID, ledgerstate.ColorIOTA, 0)

	if !clu.VerifyAddressBalances(myAddress, solo.Saldo, map[ledgerstate.Color]uint64{
		ledgerstate.ColorIOTA: solo.Saldo,
	}, "myAddress after withdraw") {
		t.Fail()
	}
}
