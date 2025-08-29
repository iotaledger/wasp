package testcore

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/wasp/v2/clients/iota-go/contracts"
	"github.com/iotaledger/wasp/v2/packages/coin"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/parameters"
	"github.com/iotaledger/wasp/v2/packages/solo"
	"github.com/iotaledger/wasp/v2/packages/testutil/testmisc"
	"github.com/iotaledger/wasp/v2/packages/vm"
	"github.com/iotaledger/wasp/v2/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/v2/packages/vm/core/evm"
	"github.com/iotaledger/wasp/v2/packages/vm/core/governance"
	"github.com/iotaledger/wasp/v2/packages/vm/gas"
)

func TestAccounts_Deposit(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{})
	sender, _ := env.NewKeyPairWithFunds(env.NewSeedFromTestNameAndTimestamp(t.Name()))
	ch := env.NewChain()

	err := ch.DepositBaseTokensToL2(100_000, sender)
	require.NoError(t, err)

	rec := ch.LastReceipt()
	require.NotNil(t, rec)
	t.Logf("========= receipt: %s", rec)
	t.Logf("========= burn log:\n%s", rec.GasBurnLog)
}

func TestAccounts_DepositWithObject(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{})
	sender, _ := env.NewKeyPairWithFunds(env.NewSeedFromTestNameAndTimestamp(t.Name()))
	ch := env.NewChain()

	obj := env.L1MintObject(sender)

	err := ch.DepositAssetsToL2(isc.NewAssets(100_000).AddObject(obj), sender)
	require.NoError(t, err)

	l2Objsecs := ch.L2Objects(isc.NewAddressAgentID(sender.Address()))
	require.Len(t, l2Objsecs, 1)
	require.EqualValues(t, obj, l2Objsecs[0])
}

// allowance shouldn't allow you to bypass gas fees.
func TestAccounts_DepositCheatAllowance(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{})
	sender, senderAddr := env.NewKeyPairWithFunds(env.NewSeedFromTestNameAndTimestamp(t.Name()))
	senderAgentID := isc.NewAddressAgentID(senderAddr)
	ch := env.NewChain()

	const baseTokensSent = 1 * isc.Million

	// send a request where allowance == assets - so that no base tokens are available outside allowance
	_, err := ch.PostRequestSync(
		solo.NewCallParams(accounts.FuncDeposit.Message()).
			AddBaseTokens(baseTokensSent).
			WithGasBudget(100_000).
			AddAllowanceBaseTokens(baseTokensSent),
		sender,
	)
	require.Error(t, err)

	rec := ch.LastReceipt()
	finalBalance := ch.L2BaseTokens(senderAgentID)
	require.Less(t, finalBalance, coin.Value(baseTokensSent))
	require.EqualValues(t, baseTokensSent, finalBalance+rec.GasFeeCharged)
}

func TestAccounts_WithdrawEverything(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{})
	sender, senderAddr := env.NewKeyPairWithFunds(env.NewSeedFromTestNameAndTimestamp(t.Name()))
	senderAgentID := isc.NewAddressAgentID(senderAddr)
	ch := env.NewChain()

	// deposit some base tokens to L2
	baseTokensToDepositToL2 := solo.BaseTokensForL2Gas
	err := ch.DepositBaseTokensToL2(baseTokensToDepositToL2, sender)
	require.NoError(t, err)

	l2balance := ch.L2BaseTokens(senderAgentID)

	// construct the request to estimate an withdrawal (leave a few tokens to pay for gas)
	req := solo.NewCallParams(accounts.FuncWithdraw.Message()).
		AddAllowance(isc.NewAssets(l2balance - solo.BaseTokensForL2Gas/10)).
		WithMaxAffordableGasBudget()

	_, estimate, err := ch.EstimateGasOffLedger(req, sender)
	require.NoError(t, err)

	// set the allowance to the maximum possible value
	req = req.WithAllowance(isc.NewAssets(l2balance - estimate.GasFeeCharged)).
		WithGasBudget(estimate.GasBurned)

	// retry the estimation (fee will be lower when writing "0" to the user account, instead of some positive number)
	_, estimate2, err := ch.EstimateGasOffLedger(req, sender)
	require.NoError(t, err)

	// set the allowance to the maximum possible value
	tokensWithdrawn := l2balance - estimate2.GasFeeCharged
	req = req.WithAllowance(isc.NewAssets(tokensWithdrawn)).
		WithGasBudget(estimate2.GasBurned)

	l1Before := ch.Env.L1BaseTokens(senderAddr)

	// withdraw all
	_, err = ch.PostRequestOffLedger(req, sender)
	require.NoError(t, err)

	require.Equal(t, estimate2.GasFeeCharged, ch.LastReceipt().GasFeeCharged)

	finalL2Balance := ch.L2BaseTokens(senderAgentID)
	require.Zero(t, finalL2Balance)

	l1After := ch.Env.L1BaseTokens(senderAddr)
	require.EqualValues(t, l1Before+tokensWithdrawn, l1After)
}

type accountsDepositTest struct {
	env               *solo.Solo
	chainAdmin        *cryptolib.KeyPair
	chainAdminAddr    *cryptolib.Address
	chainAdminAgentID isc.AgentID
	user              *cryptolib.KeyPair
	userAddr          *cryptolib.Address
	userAgentID       isc.AgentID
	ch                *solo.Chain
	req               *solo.CallParams
	coinType          coin.Type
}

func initDepositTest(t *testing.T, initCommonAccountBaseTokens ...coin.Value) *accountsDepositTest {
	ret := &accountsDepositTest{}
	ret.env = solo.New(t, &solo.InitOptions{Debug: true, PrintStackTrace: true})

	ret.chainAdmin, ret.chainAdminAddr = ret.env.NewKeyPairWithFunds(ret.env.NewSeedFromTestNameAndTimestamp(t.Name()))
	ret.chainAdminAgentID = isc.NewAddressAgentID(ret.chainAdminAddr)
	ret.user, ret.userAddr = ret.env.NewKeyPairWithFunds(ret.env.NewSeedFromTestNameAndTimestamp(t.Name()))
	ret.userAgentID = isc.NewAddressAgentID(ret.userAddr)

	initBaseTokens := coin.Value(isc.GasCoinTargetValue)
	if len(initCommonAccountBaseTokens) != 0 {
		initBaseTokens = initCommonAccountBaseTokens[0]
	}
	ret.ch, _ = ret.env.NewChainExt(ret.chainAdmin, initBaseTokens, "chain1", evm.DefaultChainID, governance.DefaultBlockKeepAmount, nil, nil)

	ret.req = solo.NewCallParams(accounts.FuncDeposit.Message())
	return ret
}

// initWithdrawTest deploys TestCoin, mints 1M tokens and deposits 100 to user's account
func initWithdrawTest(t *testing.T) *accountsDepositTest {
	v := initDepositTest(t)
	v.ch.MustDepositBaseTokensToL2(solo.BaseTokensForL2Gas, v.user)
	coinPackageID, treasuryCap := v.ch.Env.L1DeployCoinPackage(v.user)
	v.coinType = coin.MustTypeFromString(fmt.Sprintf(
		"%s::%s::%s",
		coinPackageID.String(),
		contracts.TestcoinModuleName,
		contracts.TestcoinTypeTag,
	))
	v.ch.Env.L1MintCoin(
		v.user,
		coinPackageID,
		contracts.TestcoinModuleName,
		contracts.TestcoinTypeTag,
		treasuryCap,
		1*isc.Million,
	)
	err := v.ch.DepositAssetsToL2(isc.NewEmptyAssets().AddCoin(v.coinType, 100), v.user)
	require.NoError(t, err)
	// prepare request parameters to withdraw
	// do not run the request yet
	v.req = solo.NewCallParams(accounts.FuncWithdraw.Message()).
		AddBaseTokens(solo.BaseTokensForL2Gas / 10).
		WithGasBudget(100_000)
	v.printBalances("BEGIN")
	return v
}

func (v *accountsDepositTest) printBalances(prefix string) {
	v.env.T.Logf("%s: user L1 base tokens: %d", prefix, v.env.L1BaseTokens(v.userAddr))
	v.env.T.Logf("%s: user L1 tokens: %s : %d", prefix, v.coinType, v.env.L1CoinBalance(v.userAddr, v.coinType))
	v.env.T.Logf("%s: user L2: %s", prefix, v.ch.L2Assets(v.userAgentID))
	v.env.T.Logf("%s: common account L2: %s", prefix, v.ch.L2CommonAccountAssets())
}

func TestAccounts_WithdrawDepositCoins(t *testing.T) {
	t.Run("withdraw with empty", func(t *testing.T) {
		v := initWithdrawTest(t)
		_, err := v.ch.PostRequestSync(v.req, v.user)
		testmisc.RequireErrorToBe(t, err, "not enough allowance")
	})
	t.Run("withdraw almost all", func(t *testing.T) {
		v := initWithdrawTest(t)
		toWithdraw := v.ch.L2Assets(v.userAgentID)
		t.Logf("assets to withdraw: %s", toWithdraw.String())
		// withdraw all tokens to L1
		v.req.AddAllowance(toWithdraw)
		coinsBefore := v.env.L1CoinBalance(v.userAddr, v.coinType)
		_, err := v.ch.PostRequestSync(v.req, v.user)
		require.NoError(t, err)
		v.env.AssertL1Coins(v.userAddr, v.coinType, coinsBefore+coin.Value(100))
		v.printBalances("END")
	})

	t.Run("accounting and pruning", func(t *testing.T) {
		// mint 100 tokens from chain 1 and withdraw those to L1
		v := initWithdrawTest(t)

		// create a new chain (ch2) with active state pruning set to keep only 1 block
		blockKeepAmount := int32(1)
		ch2, _ := v.env.NewChainExt(nil, 0, "evmchain", evm.DefaultChainID, blockKeepAmount, nil, nil)

		// deposit 1 native token from L1 into ch2
		err := ch2.DepositAssetsToL2(isc.NewAssets(1*isc.Million).AddCoin(v.coinType, coin.Value(1)), v.user)
		require.NoError(t, err)

		// make the chain produce 2 blocks (prune the previous block with the initial deposit info)
		for range 2 {
			_, err = ch2.PostRequestSync(solo.NewCallParamsEx("contract", "func"), v.user)
			require.Error(t, err)                      // dummy request, so an error is expected
			require.NotNil(t, ch2.LastReceipt().Error) // but it produced a receipt, thus make the state progress
		}

		// deposit 1 more after the initial deposit block has been prunned
		err = ch2.DepositAssetsToL2(isc.NewAssets(1*isc.Million).AddCoin(v.coinType, coin.Value(1)), v.user)
		require.NoError(t, err)
	})
}

func TestAccounts_TransferAndCheckBaseTokens(t *testing.T) {
	// initializes it all and prepares withdraw request, does not post it
	v := initWithdrawTest(t)
	initialCommonAccountBaseTokens := v.ch.L2CommonAccountAssets().BaseTokens()
	initialAdminAccountBaseTokens := v.ch.L2Assets(v.chainAdminAgentID).BaseTokens()

	// deposit some base tokens into the common account
	someUserWallet, _ := v.env.NewKeyPairWithFunds()
	err := v.ch.SendFromL1ToL2Account(isc.NewAssets(10*isc.Million).Coins, solo.BaseTokensForL2Gas, accounts.CommonAccount(), someUserWallet)
	require.NoError(t, err)
	commonAccBaseTokens := initialCommonAccountBaseTokens + 10*isc.Million
	require.EqualValues(t, commonAccBaseTokens, v.ch.L2CommonAccountAssets().BaseTokens())
	require.EqualValues(t, initialAdminAccountBaseTokens+v.ch.LastReceipt().GasFeeCharged, v.ch.L2Assets(v.chainAdminAgentID).BaseTokens())
	require.EqualValues(t, commonAccBaseTokens, v.ch.L2CommonAccountAssets().BaseTokens())
}

func TestAccounts_TransferPartialAssets(t *testing.T) {
	// setup a chain with some base tokens and native tokens for user1
	v := initWithdrawTest(t)
	v.ch.MustDepositBaseTokensToL2(solo.BaseTokensForL2Gas, v.ch.ChainAdmin)
	v.ch.MustDepositBaseTokensToL2(solo.BaseTokensForL2Gas, v.user)

	v.ch.AssertL2Coins(v.userAgentID, v.coinType, coin.Value(100))
	v.ch.AssertL2TotalCoins(v.coinType, coin.Value(100))

	// send funds to user2
	user2, user2Addr := v.env.NewKeyPairWithFunds(v.env.NewSeedFromTestNameAndTimestamp(t.Name()))
	user2AgentID := isc.NewAddressAgentID(user2Addr)

	// deposit 1 base token to "create account" for user2 // TODO maybe remove if account creation is not needed
	v.ch.AssertL2BaseTokens(user2AgentID, 0)
	const baseTokensToSend = 3e6
	err := v.ch.SendFromL1ToL2AccountBaseTokens(baseTokensToSend, solo.BaseTokensForL2Gas, user2AgentID, user2)
	rec := v.ch.LastReceipt()
	require.NoError(t, err)
	v.env.T.Logf("gas fee charged: %d", rec.GasFeeCharged)
	expectedUser2 := solo.BaseTokensForL2Gas + baseTokensToSend - rec.GasFeeCharged
	v.ch.AssertL2BaseTokens(user2AgentID, expectedUser2)
	// -----------------------------
	err = v.ch.SendFromL2ToL2Account(
		isc.NewAssets(baseTokensToSend).AddCoin(v.coinType, coin.Value(9)),
		user2AgentID,
		v.user,
	)
	require.NoError(t, err)

	// assert that balances are correct
	v.ch.AssertL2Coins(v.userAgentID, v.coinType, coin.Value(91))
	v.ch.AssertL2Coins(user2AgentID, v.coinType, coin.Value(9))
	v.ch.AssertL2BaseTokens(user2AgentID, expectedUser2+baseTokensToSend)
	v.ch.AssertL2TotalCoins(v.coinType, coin.Value(100))
}

func TestAccounts_DepositRandomContractMinFee(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{})
	ch := env.NewChain()

	wallet, addr := ch.Env.NewKeyPairWithFunds()
	agentID := isc.NewAddressAgentID(addr)

	var sent coin.Value = 1 * isc.Million
	_, err := ch.PostRequestSync(solo.NewCallParamsEx("", "").AddBaseTokens(sent), wallet)
	require.Error(t, err)
	receipt := ch.LastReceipt()
	require.Error(t, receipt.Error)

	require.EqualValues(t, gas.DefaultFeePolicy().MinFee(nil, parameters.BaseTokenDecimals), receipt.GasFeeCharged)
	require.EqualValues(t, sent-receipt.GasFeeCharged, ch.L2BaseTokens(agentID))
}

func TestAccounts_AllowanceNotEnoughFunds(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{})
	ch := env.NewChain()

	wallet, _ := ch.Env.NewKeyPairWithFunds()
	allowances := []*isc.Assets{
		// test base token
		isc.NewAssets(1000 * isc.Million),
		// test coins
		isc.NewEmptyAssets().AddCoin(coin.MustTypeFromString("0x2::foo::bar"), coin.Value(10)),
		// test NFTs
		// !!! TODO
		// isc.NewEmptyAssets().AddObject(iotago.Address{1, 2, 3}),
	}
	for _, a := range allowances {
		_, err := ch.PostRequestSync(
			solo.NewCallParams(accounts.FuncDeposit.Message()).
				AddBaseTokens(1*isc.Million).
				WithAllowance(a).
				WithMaxAffordableGasBudget(),
			wallet)
		require.Error(t, err)
		testmisc.RequireErrorToBe(t, err, vm.ErrNotEnoughFundsForAllowance)
		receipt := ch.LastReceipt()
		require.EqualValues(t, gas.DefaultFeePolicy().MinFee(nil, parameters.BaseTokenDecimals), receipt.GasFeeCharged)
	}
}

func TestAccounts_DepositWithNoGasBudget(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{})
	senderWallet, _ := env.NewKeyPairWithFunds(env.NewSeedFromTestNameAndTimestamp(t.Name()))
	ch := env.NewChain()

	// try to deposit with 0 gas budget
	_, err := ch.PostRequestSync(
		solo.NewCallParams(accounts.FuncDeposit.Message()).
			WithFungibleTokens(isc.NewAssets(2*isc.Million).Coins).
			WithGasBudget(0),
		senderWallet,
	)
	require.NoError(t, err)

	rec := ch.LastReceipt()
	// request should succeed, while using gas > 0, the gasBudget should be correct in the receipt
	require.Nil(t, rec.Error)
	require.NotZero(t, rec.GasBurned)
	require.EqualValues(t, ch.GetGasLimits().MinGasPerRequest, rec.GasBudget)
}

func TestAccounts_RequestWithNoGasBudget(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{})
	ch := env.NewChain()
	senderWallet, _ := env.NewKeyPairWithFunds()
	req := solo.NewCallParamsEx("dummy", "dummy").WithGasBudget(0)

	// offledger request with 0 gas
	_, err := ch.PostRequestOffLedger(req, senderWallet)
	require.EqualValues(t, 0, ch.LastReceipt().GasBudget)
	testmisc.RequireErrorToBe(t, err, vm.ErrContractNotFound)

	// post the request via on-ledger (the account has funds now), the request gets bumped to "minGasBudget"
	_, err = ch.PostRequestSync(req.WithFungibleTokens(isc.NewAssets(10*isc.Million).Coins), senderWallet)
	require.EqualValues(t, gas.LimitsDefault.MinGasPerRequest, ch.LastReceipt().GasBudget)
	testmisc.RequireErrorToBe(t, err, vm.ErrContractNotFound)

	// post the request off-ledger again (the account has funds now), the request gets bumped to "minGasBudget"
	_, err = ch.PostRequestOffLedger(req, senderWallet)
	require.EqualValues(t, gas.LimitsDefault.MinGasPerRequest, ch.LastReceipt().GasBudget)
	testmisc.RequireErrorToBe(t, err, vm.ErrContractNotFound)
}

func TestAccounts_Nonces(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{})
	ch := env.NewChain()
	senderWallet, _ := env.NewKeyPairWithFunds()
	ch.DepositAssetsToL2(isc.NewAssets(10*isc.Million), senderWallet)

	req := solo.NewCallParamsEx("dummy", "dummy").WithGasBudget(0).WithNonce(0)
	_, err := ch.PostRequestOffLedger(req, senderWallet)
	testmisc.RequireErrorToBe(t, err, vm.ErrContractNotFound)

	req = req.WithNonce(1)
	_, err = ch.PostRequestOffLedger(req, senderWallet)
	testmisc.RequireErrorToBe(t, err, vm.ErrContractNotFound)

	req = req.WithNonce(2)
	_, err = ch.PostRequestOffLedger(req, senderWallet)
	testmisc.RequireErrorToBe(t, err, vm.ErrContractNotFound)

	// try to send old nonce
	req = req.WithNonce(1)
	_, err = ch.PostRequestOffLedger(req, senderWallet)
	testmisc.RequireErrorToBe(t, err, "request was skipped")

	// try to replay nonce 2
	req = req.WithNonce(2)
	_, err = ch.PostRequestOffLedger(req, senderWallet)
	testmisc.RequireErrorToBe(t, err, "request was skipped")

	// nonce too high
	req = req.WithNonce(20)
	_, err = ch.PostRequestOffLedger(req, senderWallet)
	testmisc.RequireErrorToBe(t, err, "request was skipped")

	// correct nonce passes
	req = req.WithNonce(3)
	_, err = ch.PostRequestOffLedger(req, senderWallet)
	testmisc.RequireErrorToBe(t, err, vm.ErrContractNotFound)
}

func TestAccounts_AdjustCommonAccountBaseTokens(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{})
	sender, _ := env.NewKeyPairWithFunds(env.NewSeedFromTestNameAndTimestamp(t.Name()))
	ch := env.NewChain()

	err := ch.DepositBaseTokensToL2(10*isc.Million, sender)
	require.NoError(t, err)

	l1Bal1 := lo.Return2(ch.GetLatestAnchorWithBalances()).BaseTokens()
	l2Total1 := ch.L2TotalAssets().BaseTokens()
	require.Equal(t, l1Bal1, l2Total1)

	_, err = ch.PostRequestOffLedger(
		solo.NewCallParams(accounts.AdjustCommonAccountBaseTokens.Message(1_000, 0)).
			WithMaxAffordableGasBudget(),
		ch.ChainAdmin,
	)
	require.NoError(t, err)

	l1Bal2 := lo.Return2(ch.GetLatestAnchorWithBalances()).BaseTokens()
	l2Total2 := ch.L2TotalAssets().BaseTokens()
	require.Equal(t, l1Bal2+1000, l2Total2)

	_, err = ch.PostRequestOffLedger(
		solo.NewCallParams(accounts.AdjustCommonAccountBaseTokens.Message(0, 1_000)).
			WithMaxAffordableGasBudget(),
		ch.ChainAdmin,
	)
	require.NoError(t, err)

	l1Bal3 := lo.Return2(ch.GetLatestAnchorWithBalances()).BaseTokens()
	l2Total3 := ch.L2TotalAssets().BaseTokens()
	require.Equal(t, l1Bal3, l2Total3)
}
