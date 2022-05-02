package tests

import (
	"testing"
	"time"

	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/utxodb"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/stretchr/testify/require"
)

func TestDepositWithdraw(t *testing.T) {
	e := setupWithNoChain(t)

	chain, err := e.clu.DeployDefaultChain()
	require.NoError(t, err)
	chainNodeCount := uint64(len(chain.AllPeers))

	chEnv := newChainEnv(t, e.clu, chain)

	myWallet, myAddress, err := e.clu.NewKeyPairWithFunds()
	require.NoError(e.t, err)

	e.requestFunds(myAddress, "myAddress")
	if !e.clu.AssertAddressBalances(myAddress,
		iscp.NewTokensIotas(utxodb.FundsFromFaucetAmount)) {
		t.Fail()
	}
	if !e.clu.AssertAddressBalances(chain.OriginatorAddress(),
		iscp.NewTokensIotas(utxodb.FundsFromFaucetAmount-someIotas-1-chainNodeCount)) {
		t.Fail()
	}
	if !e.clu.AssertAddressBalances(chain.ChainAddress(),
		iscp.NewTokensIotas(someIotas+1+chainNodeCount)) {
		t.Fail()
	}
	chEnv.checkLedger()

	myAgentID := iscp.NewAgentID(myAddress, 0)
	origAgentID := iscp.NewAgentID(chain.OriginatorAddress(), 0)

	chEnv.checkBalanceOnChain(origAgentID, iscp.IotaTokenID, 0)
	chEnv.checkBalanceOnChain(myAgentID, iscp.IotaTokenID, 0)
	chEnv.checkLedger()

	// deposit some iotas to the chain
	depositIotas := uint64(42)
	chClient := chainclient.New(e.clu.L1Client(), e.clu.WaspClient(0), chain.ChainID, myWallet)

	par := chainclient.NewPostRequestParams().WithIotas(depositIotas)
	reqTx, err := chClient.Post1Request(accounts.Contract.Hname(), accounts.FuncDeposit.Hname(), *par)
	require.NoError(t, err)

	_, err = chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(chain.ChainID, reqTx, 30*time.Second)
	require.NoError(t, err)
	chEnv.checkLedger()
	chEnv.checkBalanceOnChain(myAgentID, iscp.IotaTokenID, depositIotas)
	chEnv.checkBalanceOnChain(origAgentID, iscp.IotaTokenID, 0)

	if !e.clu.AssertAddressBalances(myAddress,
		iscp.NewTokensIotas(utxodb.FundsFromFaucetAmount-depositIotas)) {
		t.Fail()
	}

	// withdraw iotas back
	reqTx3, err := chClient.Post1Request(accounts.Contract.Hname(), accounts.FuncWithdraw.Hname())
	require.NoError(t, err)
	_, err = chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(chain.ChainID, reqTx3, 30*time.Second)
	require.NoError(t, err)

	require.NoError(t, err)
	chEnv.checkLedger()
	chEnv.checkBalanceOnChain(myAgentID, iscp.IotaTokenID, 0)

	if !e.clu.AssertAddressBalances(myAddress, iscp.NewTokensIotas(utxodb.FundsFromFaucetAmount)) {
		t.Fail()
	}
}
