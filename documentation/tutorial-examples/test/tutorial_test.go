package test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/testutil/utxodb"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/corecontracts"
)

func TestTutorialFirst(t *testing.T) {
	env := solo.New(t)
	chain := env.NewChain()

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
	t.Logf("address of the user is: %s", userAddress.Bech32(parameters.L1().Protocol.Bech32HRP))
	numBaseTokens := env.L1BaseTokens(userAddress)
	t.Logf("balance of the user is: %d base tokens", numBaseTokens)
	env.AssertL1BaseTokens(userAddress, utxodb.FundsFromFaucetAmount)
}

func TestTutorialDeploySC(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: true})
	chain := env.NewChain()
	err := chain.DeployWasmContract(nil, "solotutorial", "solotutorial_bg.wasm")
	require.NoError(t, err)
}

func TestTutorialInvokeSC(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: true})
	chain := env.NewChain()
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
	require.Equal(t, "Hello, world!", codec.MustDecodeString(res.Get("str")))
}

func TestTutorialInvokeSCOffLedger(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: true})
	chain := env.NewChain()
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
	require.Equal(t, "Hello, world!", codec.MustDecodeString(res.Get("str")))
}

func TestTutorialInvokeSCError(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: true})
	chain := env.NewChain()
	err := chain.DeployWasmContract(nil, "solotutorial", "solotutorial_bg.wasm")
	require.NoError(t, err)

	// missing the required parameter "str"
	req := solo.NewCallParams("solotutorial", "storeString").
		WithMaxAffordableGasBudget()

	_, err = chain.PostRequestSync(req, nil)
	require.Error(t, err)
	require.True(t, err.Error() == "WASM: panic in VM: missing mandatory param: str")
}

func TestTutorialAccounts(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{
		AutoAdjustStorageDeposit: true,
		GasBurnLogEnabled:        true,
	})
	chain := env.NewChain()

	// create a wallet with some base tokens on L1:
	userWallet, userAddress := env.NewKeyPairWithFunds(env.NewSeedFromIndex(0))
	env.AssertL1BaseTokens(userAddress, utxodb.FundsFromFaucetAmount)

	// the wallet can we identified on L2 by an AgentID:
	userAgentID := isc.NewAgentID(userAddress)
	// for now our on-chain account is empty:
	chain.AssertL2BaseTokens(userAgentID, 0)

	// send 1 Mi from the L1 wallet to own account on-chain, controlled by the same wallet
	req := solo.NewCallParams(accounts.Contract.Name, accounts.FuncDeposit.Name).
		AddBaseTokens(1 * isc.Million).
		WithMaxAffordableGasBudget()

	// estimate the gas fee and storage deposit
	_, receipt1, err := chain.EstimateGasOnLedger(req, userWallet)
	require.NoError(t, err)
	storageDeposit1 := chain.EstimateNeededStorageDeposit(req, userWallet)
	require.Zero(t, storageDeposit1) // since 1 Mi is enough

	// send the deposit request
	req.WithGasBudget(receipt1.GasBurned).
		AddBaseTokens(receipt1.GasFeeCharged) // including base tokens for gas fee
	_, err = chain.PostRequestSync(req, userWallet)
	require.NoError(t, err)

	// our L1 balance is 1 Mi + gas fee short
	env.AssertL1BaseTokens(userAddress, utxodb.FundsFromFaucetAmount-1*isc.Million-receipt1.GasFeeCharged)
	// our L2 balance is 1 Mi
	onChainBalance := 1 * isc.Million
	chain.AssertL2BaseTokens(userAgentID, 1*isc.Million)
	// (the gas fee went to the chain's private account)

	// TODO the withdrawal part is pretty confusing for a "tutorial", this needs to be improved.

	// withdraw all base tokens back to L1
	req = solo.NewCallParams(accounts.Contract.Name, accounts.FuncWithdraw.Name).
		WithAllowance(isc.NewAssetsBaseTokens(onChainBalance - 1000)). // leave some tokens out of allowance, to pay for gas
		WithMaxAffordableGasBudget()

	// estimate the gas fee
	_, receipt2, err := chain.EstimateGasOffLedger(req, userWallet)
	require.NoError(t, err)

	// re-estimate with fixed budget and fee (the final fee might be smaller, because less gas will be charged when setting 0 in the user account, rather than a positive number)
	req.WithGasBudget(receipt2.GasBurned).
		WithAllowance(isc.NewAssetsBaseTokens(onChainBalance - (receipt2.GasFeeCharged)))
	_, receipt3, err := chain.EstimateGasOffLedger(req, userWallet)
	require.NoError(t, err)

	// send the withdraw request
	req.WithGasBudget(receipt3.GasBurned).
		WithAllowance(isc.NewAssetsBaseTokens(onChainBalance - (receipt3.GasFeeCharged)))
	_, err = chain.PostRequestOffLedger(req, userWallet)
	require.NoError(t, err)

	rec := chain.LastReceipt()
	require.EqualValues(t, rec.GasFeeCharged, receipt3.GasFeeCharged)
	require.EqualValues(t, rec.GasBurned, receipt3.GasBurned)

	// we are back to the initial situation, having been charged some gas fees
	// in the process:
	chain.AssertL2BaseTokens(userAgentID, 0)
	env.AssertL1BaseTokens(userAddress, utxodb.FundsFromFaucetAmount-receipt1.GasFeeCharged-receipt3.GasFeeCharged)
}
