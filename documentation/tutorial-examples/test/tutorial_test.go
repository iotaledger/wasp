package test

import (
	"testing"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/utxodb"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/corecontracts"
	"github.com/stretchr/testify/require"
)

func TestTutorialFirst(t *testing.T) {
	env := solo.New(t)
	chain := env.NewChain(nil, "tutorial1")

	// calls views governance::ViewGetChainInfo and root:: ViewGetContractRecords
	chainID, chainOwnerID, coreContracts := chain.GetInfo()
	// assert that all core contracts are deployed
	require.EqualValues(t, len(corecontracts.All), len(coreContracts))

	t.Logf("chain ID: %s", chainID.String())
	t.Logf("chain owner ID: %s", chainOwnerID.String())
	for hname, rec := range coreContracts {
		t.Logf("    Core contract %q: %s", rec.Name, hname)
	}
}

func TestTutorialL1(t *testing.T) {
	env := solo.New(t)
	_, userAddress := env.NewKeyPairWithFunds(env.NewSeedFromIndex(1))
	t.Logf("address of the user is: %s", userAddress.Bech32(parameters.L1.Protocol.Bech32HRP))
	numBaseTokens := env.L1BaseTokens(userAddress)
	t.Logf("balance of the user is: %d base tokens", numBaseTokens)
	env.AssertL1BaseTokens(userAddress, utxodb.FundsFromFaucetAmount)
}

func TestTutorialDeploySC(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{AutoAdjustDustDeposit: true})
	chain := env.NewChain(nil, "example")
	err := chain.DeployWasmContract(nil, "solotutorial", "solotutorial_bg.wasm")
	require.NoError(t, err)
}

func TestTutorialInvokeSC(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{AutoAdjustDustDeposit: true})
	chain := env.NewChain(nil, "ch1")
	err := chain.DeployWasmContract(nil, "solotutorial", "solotutorial_bg.wasm")
	require.NoError(t, err)

	// invoke the `storeString` function
	req := solo.NewCallParams("solotutorial", "storeString", "str", "Hello, world!").
		WithMaxAffordableGasBudget()
	_, err = chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	// invoke the `getString` view
	res, err := chain.CallView("solotutorial", "getString")
	require.NoError(t, err)
	require.Equal(t, "Hello, world!", codec.MustDecodeString(res.MustGet("str")))
}

func TestTutorialInvokeSCOffLedger(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{AutoAdjustDustDeposit: true})
	chain := env.NewChain(nil, "ch1")
	err := chain.DeployWasmContract(nil, "solotutorial", "solotutorial_bg.wasm")
	require.NoError(t, err)
	req := solo.NewCallParams("solotutorial", "storeString", "str", "Hello, world!").
		WithMaxAffordableGasBudget()

	user, _ := env.NewKeyPairWithFunds(env.NewSeedFromIndex(1))
	chain.DepositBaseTokensToL2(10_000, user)
	_, err = chain.PostRequestOffLedger(req, user)
	require.NoError(t, err)

	// invoke the `getString` view
	res, err := chain.CallView("solotutorial", "getString")
	require.NoError(t, err)
	require.Equal(t, "Hello, world!", codec.MustDecodeString(res.MustGet("str")))
}

func TestTutorialInvokeSCError(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{AutoAdjustDustDeposit: true})
	chain := env.NewChain(nil, "ch1")
	err := chain.DeployWasmContract(nil, "solotutorial", "solotutorial_bg.wasm")
	require.NoError(t, err)

	// missing the required parameter "str"
	req := solo.NewCallParams("solotutorial", "storeString").
		WithMaxAffordableGasBudget()

	_, err = chain.PostRequestSync(req, nil)
	t.Log(err)
	require.Error(t, err)
}

func TestTutorialAccounts(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{AutoAdjustDustDeposit: true})
	chain := env.NewChain(nil, "ch1")

	// create a wallet with some base tokens on L1:
	userWallet, userAddress := env.NewKeyPairWithFunds(env.NewSeedFromIndex(0))
	env.AssertL1BaseTokens(userAddress, utxodb.FundsFromFaucetAmount)

	// the wallet can we identified on L2 by an AgentID:
	userAgentID := isc.NewAgentID(userAddress)
	// for now our on-chain account is empty:
	chain.AssertL2BaseTokens(userAgentID, 0)

	// send 1 Mi from the L1 wallet to own account on-chain, controlled by the same wallet
	req := solo.NewCallParams(accounts.Contract.Name, accounts.FuncDeposit.Name).
		AddBaseTokens(1 * isc.Mi)

	// estimate the gas fee and storage deposit
	gas1, gasFee1, err := chain.EstimateGasOnLedger(req, userWallet, true)
	require.NoError(t, err)
	storageDeposit1, err := chain.EstimateNeededStorageDeposit(req, userWallet)
	require.NoError(t, err)
	require.Zero(t, storageDeposit1) // since 1 Mi is enough

	// send the deposit request
	req.WithGasBudget(gas1).
		AddBaseTokens(gasFee1) // including base tokens for gas fee
	_, err = chain.PostRequestSync(req, userWallet)
	require.NoError(t, err)

	// our L1 balance is 1 Mi + gas fee short
	env.AssertL1BaseTokens(userAddress, utxodb.FundsFromFaucetAmount-1*isc.Mi-gasFee1)
	// our L2 balance is 1 Mi
	chain.AssertL2BaseTokens(userAgentID, 1*isc.Mi)
	// (the gas fee went to the chain's private account)

	// withdraw all base tokens back to L1
	req = solo.NewCallParams(accounts.Contract.Name, accounts.FuncWithdraw.Name).
		WithAllowance(isc.NewAllowanceBaseTokens(1 * isc.Mi))

	// estimate the gas fee and storage deposit
	gas2, gasFee2, err := chain.EstimateGasOnLedger(req, userWallet, true)
	require.NoError(t, err)
	storageDeposit2, err := chain.EstimateNeededStorageDeposit(req, userWallet)
	require.NoError(t, err)

	// send the withdraw request
	req.WithGasBudget(gas2).
		AddBaseTokens(gasFee2 + storageDeposit2). // including base tokens for gas fee and storage
		AddAllowanceBaseTokens(storageDeposit2)   // and withdrawing the storage as well
	_, err = chain.PostRequestSync(req, userWallet)
	require.NoError(t, err)

	// we are back to the initial situation, having been charged some gas fees
	// in the process:
	env.AssertL1BaseTokens(userAddress, utxodb.FundsFromFaucetAmount-gasFee1-gasFee2)
	chain.AssertL2BaseTokens(userAgentID, 0)
}
