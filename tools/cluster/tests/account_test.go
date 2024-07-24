package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/clients/chainclient"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/testutil/utxodb"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/root"
)

// executed in cluster_test.go
func testBasicAccounts(t *testing.T, env *ChainEnv) {
	testAccounts(env)
}

func TestBasicAccountsNLow(t *testing.T) {
	runTest := func(tt *testing.T, n, t int) {
		e := setupWithNoChain(tt)
		chainNodes := make([]int, n)
		for i := range chainNodes {
			chainNodes[i] = i
		}
		chain, err := e.Clu.DeployChainWithDKG(chainNodes, chainNodes, uint16(t))
		require.NoError(tt, err)
		env := newChainEnv(tt, e.Clu, chain)
		testAccounts(env)
	}
	t.Run("N=1", func(tt *testing.T) { runTest(tt, 1, 1) })
	t.Run("N=2", func(tt *testing.T) { runTest(tt, 2, 2) })
	t.Run("N=3", func(tt *testing.T) { runTest(tt, 3, 3) })
	t.Run("N=4", func(tt *testing.T) { runTest(tt, 4, 3) })
}

func testAccounts(e *ChainEnv) {
	tx, err := e.Chain.DeployContract(inccounter.Contract.Name, inccounter.Contract.ProgramHash.String(), inccounter.InitParams(42))
	require.NoError(e.t, err)

	_, err = e.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(e.Chain.ChainID, tx, false, 30*time.Second)
	require.NoError(e.t, err)

	e.t.Logf("   %s: %s", root.Contract.Name, root.Contract.Hname().String())
	e.t.Logf("   %s: %s", accounts.Contract.Name, accounts.Contract.Hname().String())

	e.checkCoreContracts()

	for i := range e.Chain.CommitteeNodes {
		blockIndex, err2 := e.Chain.BlockIndex(i)
		require.NoError(e.t, err2)
		require.Greater(e.t, blockIndex, uint32(2))

		contractRegistry, err2 := e.Chain.ContractRegistry(i)
		require.NoError(e.t, err2)

		cr, ok := lo.Find(contractRegistry, func(item apiclient.ContractInfoResponse) bool {
			return item.HName == inccounter.Contract.Hname().String()
		})
		require.True(e.t, ok)

		require.EqualValues(e.t, inccounter.Contract.ProgramHash.Hex(), cr.ProgramHash)
		require.EqualValues(e.t, inccounter.Contract.Name, cr.Name)

		counterValue, err2 := e.Chain.GetCounterValue(i)
		require.NoError(e.t, err2)
		require.EqualValues(e.t, 42, counterValue)
	}

	myWallet, myAddress, err := e.Clu.NewKeyPairWithFunds()
	require.NoError(e.t, err)

	transferBaseTokens := 1 * isc.Million
	chClient := chainclient.New(e.Clu.L1Client(), e.Clu.WaspClient(0), e.Chain.ChainID, myWallet)

	par := chainclient.NewPostRequestParams().WithBaseTokens(transferBaseTokens)
	reqTx, err := chClient.PostRequest(inccounter.FuncIncCounter.Message(nil), *par)
	require.NoError(e.t, err)

	receipts, err := e.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(e.Chain.ChainID, reqTx, false, 10*time.Second)
	require.NoError(e.t, err)

	fees, err := iotago.DecodeUint64(receipts[0].GasFeeCharged)
	require.NoError(e.t, err)

	e.checkBalanceOnChain(isc.NewAgentID(myAddress), isc.BaseTokenID, transferBaseTokens-fees)

	for i := range e.Chain.CommitteeNodes {
		counterValue, err := e.Chain.GetCounterValue(i)
		require.NoError(e.t, err)
		require.EqualValues(e.t, 43, counterValue)
	}

	if !e.Clu.AssertAddressBalances(myAddress, isc.NewAssetsBaseTokensU64(utxodb.FundsFromFaucetAmount-transferBaseTokens)) {
		e.t.Fatal()
	}

	incCounterAgentID := isc.NewContractAgentID(e.Chain.ChainID, inccounter.Contract.Hname())
	e.checkBalanceOnChain(incCounterAgentID, isc.BaseTokenID, 0)
}

// executed in cluster_test.go
func testBasic2Accounts(t *testing.T, env *ChainEnv) {
	chain := env.Chain

	tx, err := chain.DeployContract(inccounter.Contract.Name, inccounter.Contract.ProgramHash.String(), inccounter.InitParams(42))
	require.NoError(t, err)

	_, err = chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(chain.ChainID, tx, false, 30*time.Second)
	require.NoError(env.t, err)

	env.checkCoreContracts()

	for _, i := range chain.CommitteeNodes {
		blockIndex, err2 := chain.BlockIndex(i)
		require.NoError(t, err2)
		require.Greater(t, blockIndex, uint32(2))

		contractRegistry, err2 := chain.ContractRegistry(i)
		require.NoError(t, err2)

		t.Logf("%+v", contractRegistry)
		cr, ok := lo.Find(contractRegistry, func(item apiclient.ContractInfoResponse) bool {
			return item.HName == inccounter.Contract.Hname().String()
		})
		require.True(t, ok)
		require.NotNil(t, cr)

		require.EqualValues(t, inccounter.Contract.ProgramHash.Hex(), cr.ProgramHash)
		require.EqualValues(t, inccounter.Contract.Name, cr.Name)

		counterValue, err2 := chain.GetCounterValue(i)
		require.NoError(t, err2)
		require.EqualValues(t, 42, counterValue)
	}

	originatorSigScheme := chain.OriginatorKeyPair
	originatorAddress := chain.OriginatorAddress()

	myWallet, myAddress, err := env.Clu.NewKeyPairWithFunds()
	require.NoError(t, err)

	transferBaseTokens := 1 * isc.Million
	myWalletClient := chainclient.New(env.Clu.L1Client(), env.Clu.WaspClient(0), chain.ChainID, myWallet)

	par := chainclient.NewPostRequestParams().WithBaseTokens(transferBaseTokens)
	reqTx, err := myWalletClient.PostRequest(inccounter.FuncIncCounter.Message(nil), *par)
	require.NoError(t, err)

	_, err = chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(chain.ChainID, reqTx, false, 30*time.Second)
	require.NoError(t, err)

	for _, i := range chain.CommitteeNodes {
		counterValue, err2 := chain.GetCounterValue(i)
		require.NoError(t, err2)
		require.EqualValues(t, 43, counterValue)
	}
	if !env.Clu.AssertAddressBalances(myAddress, isc.NewAssetsBaseTokensU64(utxodb.FundsFromFaucetAmount-transferBaseTokens)) {
		t.Fatal()
	}

	// withdraw back 500 base tokens to originator address
	fmt.Printf("\norig address from sigsheme: %s\n", originatorAddress.String())
	origL1Balance := env.Clu.AddressBalances(originatorAddress).BaseTokens
	originatorClient := chainclient.New(env.Clu.L1Client(), env.Clu.WaspClient(0), chain.ChainID, originatorSigScheme)
	allowanceBaseTokens := uint64(800_000)
	req2, err := originatorClient.PostOffLedgerRequest(context.Background(), accounts.FuncWithdraw.Message(),
		chainclient.PostRequestParams{
			Allowance: isc.NewAssetsBaseTokensU64(allowanceBaseTokens),
		},
	)
	require.NoError(t, err)

	_, err = chain.CommitteeMultiClient().WaitUntilRequestProcessedSuccessfully(chain.ChainID, req2.ID(), true, 30*time.Second)
	require.NoError(t, err)

	require.Equal(t, env.Clu.AddressBalances(originatorAddress).BaseTokens, origL1Balance+allowanceBaseTokens)
}
