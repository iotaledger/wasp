package testcore

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/iota-go/contracts"
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/testutil/testmisc"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

const BaseTokensDepositFee = 100

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
	baseTokensToDepositToL2 := coin.Value(100_000)
	err := ch.DepositBaseTokensToL2(baseTokensToDepositToL2, sender)
	require.NoError(t, err)

	l2balance := ch.L2BaseTokens(senderAgentID)

	// construct the request to estimate an withdrawal (leave a few tokens to pay for gas)
	req := solo.NewCallParams(accounts.FuncWithdraw.Message()).
		AddAllowance(isc.NewAssets(l2balance - 1000)).
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

// func TestAccounts_Foundries(t *testing.T) {
// 	var env *solo.Solo
// 	var ch *solo.Chain
// 	var senderKeyPair *cryptolib.KeyPair
// 	var senderAddr *cryptolib.Address
// 	var senderAgentID isc.AgentID

// 	initTest := func() {
// 		env = solo.New(t, &solo.InitOptions{})
// 		ch, _ = env.NewChainExt(nil, 10*isc.Million, "chain1", evm.DefaultChainID, governance.DefaultBlockKeepAmount)
// 		senderKeyPair, senderAddr = env.NewKeyPairWithFunds(env.NewSeedFromTestNameAndTimestamp(t.Name()))
// 		senderAgentID = isc.NewAddressAgentID(senderAddr)

// 		ch.MustDepositBaseTokensToL2(10*isc.Million, senderKeyPair)
// 	}
// 	t.Run("newFoundry fails when no allowance is provided", func(t *testing.T) {
// 		env = solo.New(t, &solo.InitOptions{})
// 		ch, _ = env.NewChainExt(nil, 100_000, "chain1", evm.DefaultChainID, governance.DefaultBlockKeepAmount)

// 		var ts iotago.TokenScheme = &iotago.SimpleTokenScheme{MaximumSupply: coin.Value(1), MintedTokens: coin.Zero, MeltedTokens: coin.Zero}
// 		req := solo.NewCallParams(accounts.FuncNativeTokenCreate.Message(
// 			isc.NewIRC30NativeTokenMetadata("TEST", "TEST", 8),
// 			&ts,
// 		)).AddBaseTokens(2 * isc.Million).WithGasBudget(math.MaxUint64)
// 		_, err := ch.PostRequestSync(req, nil)
// 		require.Error(t, err)
// 		// it succeeds when allowance is added
// 		_, err = ch.PostRequestSync(req.AddAllowanceBaseTokens(1*isc.Million), nil)
// 		require.NoError(t, err)
// 	})

// 	t.Run("newFoundry overrides bad melted/minted token counters in tokenscheme", func(t *testing.T) {
// 		env = solo.New(t, &solo.InitOptions{})
// 		ch, _ = env.NewChainExt(nil, 100_000, "chain1", evm.DefaultChainID, governance.DefaultBlockKeepAmount)

// 		var ts iotago.TokenScheme = &iotago.SimpleTokenScheme{MaximumSupply: coin.Value(1), MintedTokens: coin.Value(10), MeltedTokens: coin.Value(10)}
// 		req := solo.NewCallParams(accounts.FuncNativeTokenCreate.Message(
// 			isc.NewIRC30NativeTokenMetadata("TEST", "TEST", 8),
// 			&ts,
// 		)).AddBaseTokens(2 * isc.Million).WithGasBudget(math.MaxUint64)
// 		_, err := ch.PostRequestSync(req.AddAllowanceBaseTokens(1*isc.Million), nil)
// 		require.NoError(t, err)
// 	})
// 	t.Run("supply 10", func(t *testing.T) {
// 		initTest()
// 		sn, _, err := ch.NewNativeTokenParams(coin.Value(10)).
// 			WithUser(senderKeyPair).
// 			CreateFoundry()
// 		require.NoError(t, err)
// 		require.EqualValues(t, 1, int(sn))
// 	})
// 	t.Run("supply 1", func(t *testing.T) {
// 		initTest()
// 		sn, _, err := ch.NewNativeTokenParams(coin.Value(1)).
// 			WithUser(senderKeyPair).
// 			CreateFoundry()
// 		require.NoError(t, err)
// 		require.EqualValues(t, 1, sn)
// 	})
// 	t.Run("supply 0", func(t *testing.T) {
// 		initTest()
// 		_, _, err := ch.NewNativeTokenParams(coin.Value(0)).
// 			WithUser(senderKeyPair).
// 			CreateFoundry()
// 		testmisc.RequireErrorToBe(t, err, vm.ErrCreateFoundryMaxSupplyMustBePositive)
// 	})
// 	t.Run("supply max possible", func(t *testing.T) {
// 		initTest()
// 		sn, _, err := ch.NewNativeTokenParams(coin.MaxValue).
// 			WithUser(senderKeyPair).
// 			CreateFoundry()
// 		require.NoError(t, err)
// 		require.EqualValues(t, 1, sn)
// 	})
// 	t.Run("max supply 10, mintTokens 5", func(t *testing.T) {
// 		initTest()
// 		sn, nativeTokenID, err := ch.NewNativeTokenParams(coin.Value(10)).
// 			WithUser(senderKeyPair).
// 			CreateFoundry()
// 		require.NoError(t, err)
// 		require.EqualValues(t, 1, sn)
// 		ch.AssertL2Coins(senderAgentID, nativeTokenID, coin.Zero)
// 		ch.AssertL2TotalCoins(nativeTokenID, coin.Zero)

// 		err = ch.SendFromL1ToL2AccountBaseTokens(BaseTokensDepositFee, 1000, accounts.CommonAccount(), senderKeyPair)
// 		require.NoError(t, err)
// 		t.Logf("common account base tokens = %d before mint", ch.L2CommonAccountBaseTokens())

// 		err = ch.MintTokens(sn, coin.Value(5), senderKeyPair)
// 		require.NoError(t, err)

// 		ch.AssertL2Coins(senderAgentID, nativeTokenID, coin.Value(5))
// 		ch.AssertL2TotalCoins(nativeTokenID, coin.Value(5))

// 		testdbhash.VerifyContractStateHash(env, accounts.Contract, "", t.Name())
// 	})
// 	t.Run("max supply 1, mintTokens 1", func(t *testing.T) {
// 		initTest()
// 		sn, nativeTokenID, err := ch.NewNativeTokenParams(coin.Value(1)).
// 			WithUser(senderKeyPair).
// 			CreateFoundry()
// 		require.NoError(t, err)
// 		require.EqualValues(t, 1, sn)
// 		ch.AssertL2Coins(senderAgentID, nativeTokenID, coin.Zero)
// 		ch.AssertL2TotalCoins(nativeTokenID, coin.Zero)

// 		err = ch.SendFromL1ToL2AccountBaseTokens(BaseTokensDepositFee, 1000, accounts.CommonAccount(), senderKeyPair)
// 		require.NoError(t, err)
// 		err = ch.MintTokens(sn, coin.Value(1), senderKeyPair)
// 		require.NoError(t, err)

// 		ch.AssertL2Coins(senderAgentID, nativeTokenID, coin.Value(1))
// 		ch.AssertL2TotalCoins(nativeTokenID, coin.Value(1))
// 	})

// 	t.Run("max supply 1, mintTokens 2", func(t *testing.T) {
// 		initTest()
// 		sn, nativeTokenID, err := ch.NewNativeTokenParams(coin.Value(1)).
// 			WithUser(senderKeyPair).
// 			CreateFoundry()
// 		require.NoError(t, err)
// 		require.EqualValues(t, 1, sn)

// 		err = ch.MintTokens(sn, coin.Value(2), senderKeyPair)
// 		testmisc.RequireErrorToBe(t, err, vm.ErrNativeTokenSupplyOutOffBounds)

// 		ch.AssertL2Coins(senderAgentID, nativeTokenID, coin.Zero)
// 		ch.AssertL2TotalCoins(nativeTokenID, coin.Zero)
// 		// })
// 		t.Run("max supply 1000, mintTokens 500_500_1", func(t *testing.T) {
// 			initTest()
// 			sn, nativeTokenID, err := ch.NewNativeTokenParams(coin.Value(1000)).
// 				WithUser(senderKeyPair).
// 				CreateFoundry()
// 			require.NoError(t, err)
// 			require.EqualValues(t, 1, sn)

// 			err = ch.SendFromL1ToL2AccountBaseTokens(BaseTokensDepositFee, 1000, accounts.CommonAccount(), senderKeyPair)
// 			require.NoError(t, err)
// 			err = ch.MintTokens(sn, coin.Value(500), senderKeyPair)
// 			require.NoError(t, err)
// 			ch.AssertL2Coins(senderAgentID, nativeTokenID, coin.Value(500))
// 			ch.AssertL2TotalCoins(nativeTokenID, coin.Value(500))

// 			err = ch.MintTokens(sn, coin.Value(500), senderKeyPair)
// 			require.NoError(t, err)
// 			ch.AssertL2Coins(senderAgentID, nativeTokenID, coin.Value(1000))
// 			ch.AssertL2TotalCoins(nativeTokenID, coin.Value(1000))

// 			err = ch.MintTokens(sn, coin.Value(1), senderKeyPair)
// 			testmisc.RequireErrorToBe(t, err, vm.ErrNativeTokenSupplyOutOffBounds)

// 			ch.AssertL2Coins(senderAgentID, nativeTokenID, coin.Value(1000))
// 			ch.AssertL2TotalCoins(nativeTokenID, coin.Value(1000))
// 		})
// 		t.Run("max supply MaxUint256, mintTokens MaxUint256_1", func(t *testing.T) {
// 			initTest()
// 			sn, nativeTokenID, err := ch.NewNativeTokenParams(coin.MaxValue).
// 				WithUser(senderKeyPair).
// 				CreateFoundry()
// 			require.NoError(t, err)
// 			require.EqualValues(t, 1, sn)

// 			err = ch.SendFromL1ToL2AccountBaseTokens(BaseTokensDepositFee, 1000, accounts.CommonAccount(), senderKeyPair)
// 			require.NoError(t, err)
// 			err = ch.MintTokens(sn, coin.MaxValue, senderKeyPair)
// 			require.NoError(t, err)
// 			ch.AssertL2Coins(senderAgentID, nativeTokenID, coin.MaxValue)

// 			err = ch.MintTokens(sn, coin.Value(1), senderKeyPair)
// 			testmisc.RequireErrorToBe(t, err, vm.ErrOverflow)

// 			ch.AssertL2Coins(senderAgentID, nativeTokenID, coin.MaxValue)
// 			ch.AssertL2TotalCoins(nativeTokenID, coin.MaxValue)
// 		})
// 		t.Run("max supply 100, destroy fail", func(t *testing.T) {
// 			initTest()
// 			sn, nativeTokenID, err := ch.NewNativeTokenParams(coin.MaxValue).
// 				WithUser(senderKeyPair).
// 				CreateFoundry()
// 			require.NoError(t, err)
// 			require.EqualValues(t, 1, sn)

// 			err = ch.DestroyTokensOnL2(nativeTokenID, coin.Value(1), senderKeyPair)
// 			testmisc.RequireErrorToBe(t, err, accounts.ErrNotEnoughFunds)
// 			ch.AssertL2Coins(senderAgentID, nativeTokenID, coin.Zero)
// 			ch.AssertL2TotalCoins(nativeTokenID, coin.Zero)
// 		})
// 		t.Run("max supply 100, mint_20, destroy_10", func(t *testing.T) {
// 			initTest()
// 			sn, nativeTokenID, err := ch.NewNativeTokenParams(coin.Value(100)).
// 				WithUser(senderKeyPair).
// 				CreateFoundry()
// 			require.NoError(t, err)
// 			require.EqualValues(t, 1, sn)

// 			out, err := ch.GetFoundryOutput(1)
// 			require.NoError(t, err)
// 			require.EqualValues(t, out.MustNativeTokenID(), nativeTokenID)
// 			ch.AssertL2Coins(senderAgentID, nativeTokenID, coin.Zero)
// 			ch.AssertL2TotalCoins(nativeTokenID, coin.Zero)

// 			err = ch.SendFromL1ToL2AccountBaseTokens(BaseTokensDepositFee, 1000, accounts.CommonAccount(), senderKeyPair)
// 			require.NoError(t, err)
// 			err = ch.MintTokens(sn, coin.Value(20), senderKeyPair)
// 			require.NoError(t, err)
// 			ch.AssertL2Coins(senderAgentID, nativeTokenID, coin.Value(20))
// 			ch.AssertL2TotalCoins(nativeTokenID, coin.Value(20))

// 			err = ch.DestroyTokensOnL2(nativeTokenID, coin.Value(10), senderKeyPair)
// 			require.NoError(t, err)
// 			ch.AssertL2TotalCoins(nativeTokenID, coin.Value(10))
// 			ch.AssertL2Coins(senderAgentID, nativeTokenID, coin.Value(10))
// 		})
// 		t.Run("max supply 1000000, mint_1000000, destroy_1000000", func(t *testing.T) {
// 			initTest()
// 			sn, nativeTokenID, err := ch.NewNativeTokenParams(coin.Value(1_000_000)).
// 				WithUser(senderKeyPair).
// 				CreateFoundry()
// 			require.NoError(t, err)
// 			require.EqualValues(t, 1, sn)

// 			out, err := ch.GetFoundryOutput(1)
// 			require.NoError(t, err)
// 			require.EqualValues(t, out.MustNativeTokenID(), nativeTokenID)
// 			ch.AssertL2Coins(senderAgentID, nativeTokenID, coin.Value(0))
// 			ch.AssertL2TotalCoins(nativeTokenID, coin.Value(0))

// 			err = ch.SendFromL1ToL2AccountBaseTokens(BaseTokensDepositFee, 1000, accounts.CommonAccount(), senderKeyPair)
// 			require.NoError(t, err)
// 			err = ch.MintTokens(sn, coin.Value(1_000_000), senderKeyPair)
// 			require.NoError(t, err)
// 			ch.AssertL2Coins(senderAgentID, nativeTokenID, coin.Value(1_000_000))
// 			ch.AssertL2TotalCoins(nativeTokenID, coin.Value(1_000_000))
// 			out, err = ch.GetFoundryOutput(1)
// 			require.NoError(t, err)
// 			ts := util.MustTokenScheme(out.TokenScheme)
// 			require.True(t, coin.Value(1_000_000).Cmp(ts.MintedTokens) == 0)

// 			// FIXME bug iotago can't destroy foundry
// 			// err = destroyTokens(sn, coin.Value(1000000))
// 			// require.NoError(t, err)
// 			// ch.AssertL2TotalCoins(nativeTokenID, coin.Zero)
// 			// ch.AssertL2Coins(userAgentID, nativeTokenID, coin.Zero)
// 			// out, err = ch.GetFoundryOutput(1)
// 			// require.NoError(t, err)
// 			// require.True(t, coin.Zero.Cmp(out.MintedTokens) == 0)
// 		})
// 		t.Run("10 foundries", func(t *testing.T) {
// 			initTest()
// 			ch.MustDepositBaseTokensToL2(50_000_000, senderKeyPair)
// 			nativeTokenIDs := make([]isc.NativeTokenID, 11)
// 			for sn := uint32(1); sn <= 10; sn++ {
// 				snBack, nativeTokenID, err := ch.NewNativeTokenParams(coin.Value(int64(sn + 1))).
// 					WithUser(senderKeyPair).
// 					CreateFoundry()
// 				nativeTokenIDs[sn] = nativeTokenID
// 				require.NoError(t, err)
// 				require.EqualValues(t, int(sn), int(snBack))
// 				ch.AssertL2Coins(senderAgentID, nativeTokenID, coin.Zero)
// 				ch.AssertL2TotalCoins(nativeTokenID, coin.Zero)
// 			}
// 			// mint max supply from each
// 			ch.MustDepositBaseTokensToL2(50_000_000, senderKeyPair)
// 			for sn := uint32(1); sn <= 10; sn++ {
// 				err := ch.MintTokens(sn, coin.Value(int64(sn+1)), senderKeyPair)
// 				require.NoError(t, err)

// 				out, err := ch.GetFoundryOutput(sn)
// 				require.NoError(t, err)

// 				require.EqualValues(t, sn, out.SerialNumber)
// 				ts := util.MustTokenScheme(out.TokenScheme)
// 				require.True(t, ts.MaximumSupply.Cmp(coin.Value(int64(sn+1))) == 0)
// 				require.True(t, ts.MintedTokens.Cmp(coin.Value(int64(sn+1))) == 0)
// 				nativeTokenID := out.MustNativeTokenID()

// 				ch.AssertL2Coins(senderAgentID, nativeTokenID, coin.Value(int64(sn+1)))
// 				ch.AssertL2TotalCoins(nativeTokenID, coin.Value(int64(sn+1)))
// 			}
// 			// destroy 1 token of each nativeTokenID
// 			for sn := uint32(1); sn <= 10; sn++ {
// 				err := ch.DestroyTokensOnL2(nativeTokenIDs[sn], coin.Value(1), senderKeyPair)
// 				require.NoError(t, err)
// 			}
// 			// check balances
// 			for sn := uint32(1); sn <= 10; sn++ {
// 				out, err := ch.GetFoundryOutput(sn)
// 				require.NoError(t, err)

// 				require.EqualValues(t, sn, out.SerialNumber)
// 				ts := util.MustTokenScheme(out.TokenScheme)
// 				require.True(t, ts.MaximumSupply.Cmp(coin.Value(int64(sn+1))) == 0)
// 				require.True(t, coin.Value(0).Sub(ts.MintedTokens, ts.MeltedTokens).Cmp(coin.Value(int64(sn))) == 0)
// 				nativeTokenID := out.MustNativeTokenID()

// 				ch.AssertL2Coins(senderAgentID, nativeTokenID, coin.Value(int64(sn)))
// 				ch.AssertL2TotalCoins(nativeTokenID, coin.Value(int64(sn)))
// 			}
// 		})
// 	})
// 	t.Run("constant storage deposit to hold a token UTXO", func(t *testing.T) {
// 		initTest()
// 		// create a foundry for the maximum amount of tokens possible
// 		sn, nativeTokenID, err := ch.NewNativeTokenParams(coin.MaxValue).
// 			WithUser(senderKeyPair).
// 			CreateFoundry()
// 		require.NoError(t, err)

// 		err = ch.SendFromL1ToL2AccountBaseTokens(BaseTokensDepositFee, 1, accounts.CommonAccount(), senderKeyPair)
// 		require.NoError(t, err)
// 		x := ch.L2CommonAccountBaseTokens()
// 		t.Logf("common account base tokens = %d before mint", x)

// 		big1 := coin.Value(1)
// 		err = ch.MintTokens(sn, big1, senderKeyPair)
// 		require.NoError(t, err)

// 		ch.AssertL2Coins(senderAgentID, nativeTokenID, big1)
// 		ch.AssertL2TotalCoins(nativeTokenID, big1)
// 		ownerBal1 := ch.L2Assets(ch.OwnerAgentID())
// 		commonAccountBalanceBeforeLastMint := ch.L2CommonAccountBaseTokens()

// 		// after minting 1 token, try to mint the remaining tokens
// 		allOtherTokens := new(big.Int).Set(coin.MaxValue)
// 		allOtherTokens = allOtherTokens.Sub(allOtherTokens, big1)

// 		err = ch.MintTokens(sn, allOtherTokens, senderKeyPair)
// 		require.NoError(t, err)

// 		commonAccountBalanceAfterLastMint := ch.L2CommonAccountBaseTokens()
// 		require.Equal(t, commonAccountBalanceAfterLastMint, commonAccountBalanceBeforeLastMint)
// 		// assert that no extra base tokens were used for the storage deposit
// 		ownerBal2 := ch.L2Assets(ch.OwnerAgentID())
// 		receipt := ch.LastReceipt()
// 		require.Equal(t, ownerBal1.BaseTokens+receipt.GasFeeCharged, ownerBal2.BaseTokens)
// 	})
// 	t.Run("newFoundry exposes foundry serial number in event", func(t *testing.T) {
// 		initTest()
// 		sn, _, err := ch.NewNativeTokenParams(coin.MaxValue).
// 			WithUser(senderKeyPair).
// 			CreateFoundry()
// 		require.NoError(t, err)
// 		require.EqualValues(t, 1, sn)

// 		events, err := ch.GetEventsForContract(accounts.Contract.Name)
// 		require.NoError(t, err)
// 		require.Len(t, events, 1)
// 		sn, err = codec.Decode[uint32](events[0].Payload)
// 		require.NoError(t, err)
// 		require.EqualValues(t, 1, sn)
// 	})
// }

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
	ret.ch, _ = ret.env.NewChainExt(ret.chainAdmin, initBaseTokens, "chain1", evm.DefaultChainID, governance.DefaultBlockKeepAmount)

	ret.req = solo.NewCallParams(accounts.FuncDeposit.Message())
	return ret
}

// initWithdrawTest deploys TestCoin, mints 1M tokens and deposits 100 to user's account
func initWithdrawTest(t *testing.T, initCommonAccountBaseTokens ...coin.Value) *accountsDepositTest {
	v := initDepositTest(t, initCommonAccountBaseTokens...)
	v.ch.MustDepositBaseTokensToL2(2*isc.Million, v.user)
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
		AddBaseTokens(12000).
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
	t.Run("mint withdraw destroy fail", func(t *testing.T) {
		t.Skip("TODO")
		v := initWithdrawTest(t)
		allSenderAssets := v.ch.L2Assets(v.userAgentID)
		v.req.AddAllowance(allSenderAssets)
		_, err := v.ch.PostRequestSync(v.req, v.user)
		require.NoError(t, err)
		v.printBalances("AFTER MINT")
		// should fail because those tokens are not on the user's on chain account
		err = v.ch.DestroyTokensOnL2(v.coinType, coin.Value(50), v.user)
		testmisc.RequireErrorToBe(t, err, accounts.ErrNotEnoughFunds)
		v.env.AssertL1Coins(v.userAddr, v.coinType, coin.Value(100))
		v.printBalances("AFTER DESTROY")
	})
	t.Run("mint withdraw destroy success 1", func(t *testing.T) {
		t.Skip("TODO")
		v := initWithdrawTest(t)

		allSenderAssets := v.ch.L2Assets(v.userAgentID)
		v.req.AddAllowance(allSenderAssets)
		_, err := v.ch.PostRequestSync(v.req, v.user)
		require.NoError(t, err)
		v.printBalances("AFTER MINT")
		v.ch.AssertL2Coins(v.userAgentID, v.coinType, coin.Value(0))

		err = v.ch.DepositAssetsToL2(isc.NewEmptyAssets().AddCoin(v.coinType, coin.Value(50)), v.user)
		require.NoError(t, err)
		v.env.AssertL1Coins(v.userAddr, v.coinType, coin.Value(50))
		v.ch.AssertL2Coins(v.userAgentID, v.coinType, coin.Value(50))
		v.ch.AssertL2TotalCoins(v.coinType, coin.Value(50))
		v.printBalances("AFTER DEPOSIT")

		err = v.ch.DestroyTokensOnL2(v.coinType, coin.Value(49), v.user)
		require.NoError(t, err)
		v.ch.AssertL2Coins(v.userAgentID, v.coinType, coin.Value(1))
		v.env.AssertL1Coins(v.userAddr, v.coinType, coin.Value(50))
		v.printBalances("AFTER DESTROY")

		// sent the last 50 tokens to an evm account
		_, someEthereumAddr := solo.NewEthereumAccount()
		someEthereumAgentID := isc.NewEthereumAddressAgentID(someEthereumAddr)

		err = v.ch.TransferAllowanceTo(isc.NewEmptyAssets().AddCoin(v.coinType, coin.Value(50)),
			someEthereumAgentID,
			v.user,
		)
		require.NoError(t, err)
		v.ch.AssertL2Coins(v.userAgentID, v.coinType, coin.Value(1))
		v.env.AssertL1Coins(v.userAddr, v.coinType, coin.Value(0))
		v.ch.AssertL2Coins(someEthereumAgentID, v.coinType, coin.Value(50))
	})
	t.Run("unwrap use case", func(t *testing.T) {
		t.Skip("TODO")
		v := initWithdrawTest(t)
		allSenderAssets := v.ch.L2Assets(v.userAgentID)
		v.req.AddAllowance(allSenderAssets)
		_, err := v.ch.PostRequestSync(v.req, v.user)
		require.NoError(t, err)
		v.printBalances("AFTER MINT")
		v.env.AssertL1Coins(v.userAddr, v.coinType, coin.Value(100))
		v.ch.AssertL2Coins(v.userAgentID, v.coinType, coin.Value(0))

		err = v.ch.DepositAssetsToL2(isc.NewEmptyAssets().AddCoin(v.coinType, coin.Value(1)), v.user)
		require.NoError(t, err)
		v.printBalances("AFTER DEPOSIT 1")

		err = v.ch.DestroyTokensOnL1(v.coinType, coin.Value(49), v.user)
		require.NoError(t, err)
		v.printBalances("AFTER DESTROY")
		v.ch.AssertL2Coins(v.userAgentID, v.coinType, coin.Value(1))
		v.env.AssertL1Coins(v.userAddr, v.coinType, coin.Value(50))
	})
	t.Run("unwrap use case 2", func(t *testing.T) {
		t.Skip("TODO")
		v := initWithdrawTest(t)
		allSenderAssets := v.ch.L2Assets(v.userAgentID)
		v.req.AddAllowance(allSenderAssets)
		_, err := v.ch.PostRequestSync(v.req, v.user)
		require.NoError(t, err)
		v.printBalances("AFTER MINT")
		// no tokens on chain
		v.env.AssertL1Coins(v.userAddr, v.coinType, coin.Value(100))
		v.ch.AssertL2Coins(v.userAgentID, v.coinType, coin.Value(0))

		// deposit and destroy on the same req (chain currently doesn't have an internal UTXO for this tokenID)
		err = v.ch.DestroyTokensOnL1(v.coinType, coin.Value(49), v.user)
		require.NoError(t, err)
		v.printBalances("AFTER DESTROY")
		v.ch.AssertL2Coins(v.userAgentID, v.coinType, coin.Value(0))
		v.env.AssertL1Coins(v.userAddr, v.coinType, coin.Value(51))
	})
	t.Run("mint withdraw destroy fail", func(t *testing.T) {
		t.Skip("TODO")
		v := initWithdrawTest(t)
		allSenderAssets := v.ch.L2Assets(v.userAgentID)
		v.req.AddAllowance(allSenderAssets)
		_, err := v.ch.PostRequestSync(v.req, v.user)
		require.NoError(t, err)

		v.printBalances("AFTER MINT")
		v.env.AssertL1Coins(v.userAddr, v.coinType, coin.Value(100))
		v.ch.AssertL2Coins(v.userAgentID, v.coinType, coin.Value(0))

		err = v.ch.DepositAssetsToL2(isc.NewEmptyAssets().AddCoin(v.coinType, coin.Value(50)), v.user)
		require.NoError(t, err)
		v.env.AssertL1Coins(v.userAddr, v.coinType, coin.Value(50))
		v.ch.AssertL2Coins(v.userAgentID, v.coinType, coin.Value(50))
		v.ch.AssertL2TotalCoins(v.coinType, coin.Value(50))
		v.printBalances("AFTER DEPOSIT")

		err = v.ch.DestroyTokensOnL2(v.coinType, coin.Value(50), v.user)
		require.NoError(t, err)
		v.ch.AssertL2Coins(v.userAgentID, v.coinType, coin.Value(0))
		v.env.AssertL1Coins(v.userAddr, v.coinType, coin.Value(50))
	})

	t.Run("accounting and pruning", func(t *testing.T) {
		// mint 100 tokens from chain 1 and withdraw those to L1
		v := initWithdrawTest(t)

		// create a new chain (ch2) with active state pruning set to keep only 1 block
		blockKeepAmount := int32(1)
		ch2, _ := v.env.NewChainExt(nil, 0, "evmchain", evm.DefaultChainID, blockKeepAmount)

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
	err := v.ch.SendFromL1ToL2Account(11*isc.Million, isc.NewAssets(10*isc.Million).Coins, accounts.CommonAccount(), someUserWallet)
	require.NoError(t, err)
	commonAccBaseTokens := initialCommonAccountBaseTokens + 10*isc.Million
	require.EqualValues(t, commonAccBaseTokens, v.ch.L2CommonAccountAssets().BaseTokens())
	require.EqualValues(t, initialAdminAccountBaseTokens+v.ch.LastReceipt().GasFeeCharged, v.ch.L2Assets(v.chainAdminAgentID).BaseTokens())
	require.EqualValues(t, commonAccBaseTokens, v.ch.L2CommonAccountAssets().BaseTokens())
}

// func TestAccounts_FoundryDestroy(t *testing.T) {
// 	t.Run("destroy existing", func(t *testing.T) {
// 		v := initDepositTest(t, nil)
// 		v.ch.MustDepositBaseTokensToL2(2*isc.Million, v.user)
// 		sn, _, err := v.ch.NewNativeTokenParams(coin.Value(1_000_000)).
// 			WithUser(v.user).
// 			CreateFoundry()
// 		require.NoError(t, err)

// 		err = v.ch.DestroyFoundry(sn, v.user)
// 		require.NoError(t, err)
// 		_, err = v.ch.GetFoundryOutput(sn)
// 		testmisc.RequireErrorToBe(t, err, "not found")
// 	})
// 	t.Run("destroy fail", func(t *testing.T) {
// 		v := initDepositTest(t, nil)
// 		err := v.ch.DestroyFoundry(2, v.user)
// 		testmisc.RequireErrorToBe(t, err, "unauthorized")
// 	})
// }

func TestAccounts_TransferPartialAssets(t *testing.T) {
	// setup a chain with some base tokens and native tokens for user1
	v := initWithdrawTest(t)
	v.ch.MustDepositBaseTokensToL2(10*isc.Million, v.ch.ChainAdmin)
	v.ch.MustDepositBaseTokensToL2(10*isc.Million, v.user)

	v.ch.AssertL2Coins(v.userAgentID, v.coinType, coin.Value(100))
	v.ch.AssertL2TotalCoins(v.coinType, coin.Value(100))

	// send funds to user2
	user2, user2Addr := v.env.NewKeyPairWithFunds(v.env.NewSeedFromTestNameAndTimestamp(t.Name()))
	user2AgentID := isc.NewAddressAgentID(user2Addr)

	// deposit 1 base token to "create account" for user2 // TODO maybe remove if account creation is not needed
	v.ch.AssertL2BaseTokens(user2AgentID, 0)
	const baseTokensToSend = 3 * isc.Million
	err := v.ch.SendFromL1ToL2AccountBaseTokens(BaseTokensDepositFee, baseTokensToSend, user2AgentID, user2)
	rec := v.ch.LastReceipt()
	require.NoError(t, err)
	v.env.T.Logf("gas fee charged: %d", rec.GasFeeCharged)
	expectedUser2 := BaseTokensDepositFee + baseTokensToSend - rec.GasFeeCharged
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

func TestAccounts_NFTAccount(t *testing.T) {
	t.Skip("!!! TODO")
	// env := solo.New(t, &solo.InitOptions{})
	// ch := env.NewChain()
	//
	// issuerWallet, _ := ch.Env.NewKeyPairWithFunds()
	// ownerWallet, ownerAddress := ch.Env.NewKeyPairWithFunds()
	// ownerBalance := ch.Env.L1BaseTokens(ownerAddress)
	//
	// nftInfo, err := ch.Env.MintNFTL1(issuerWallet, ownerAddress, []byte("foobar"))
	// require.NoError(t, err)
	// nftAddress := nftInfo.Issuer
	//
	// // deposit funds on behalf of the NFT
	// const baseTokensToSend = 10 * isc.Million
	// req := solo.NewCallParams(accounts.FuncDeposit.Message()).
	// 	AddBaseTokens(baseTokensToSend).
	// 	WithMaxAffordableGasBudget().
	// 	WithSender(nftAddress)
	//
	// _, err = ch.PostRequestSync(req, ownerWallet)
	// require.NoError(t, err)
	// rec := ch.LastReceipt()
	//
	// nftAgentID := isc.NewAddressAgentID(nftAddress)
	// ch.AssertL2BaseTokens(nftAgentID, baseTokensToSend-rec.GasFeeCharged)
	// ch.Env.AssertL1BaseTokens(nftAddress, 0)
	// ch.Env.AssertL1BaseTokens(
	// 	ownerAddress,
	// 	ownerBalance+nftInfo.Output.Deposit()-baseTokensToSend,
	// )
	// require.True(t, ch.Env.HasL1NFT(ownerAddress, nftInfo.ID))
	//
	// // withdraw to the NFT on L1
	// const baseTokensToWithdrawal = 1 * isc.Million
	// wdReq := solo.NewCallParams(accounts.FuncWithdraw.Message()).
	// 	AddAllowanceBaseTokens(baseTokensToWithdrawal).
	// 	WithMaxAffordableGasBudget()
	//
	// // NFT owner on L1 can't move L2 funds owned by the NFT unless the request is sent in behalf of the NFT (NFTID is specified as "Sender")
	// _, err = ch.PostRequestSync(wdReq, ownerWallet)
	// require.Error(t, err)
	//
	// // NFT owner can withdraw funds owned by the NFT on the chain
	// _, err = ch.PostRequestSync(wdReq.WithSender(nftAddress), ownerWallet)
	// require.NoError(t, err)
	// ch.Env.AssertL1BaseTokens(nftAddress, baseTokensToWithdrawal)
}

func checkChainNFTData(t *testing.T, ch *solo.Chain, obj isc.IotaObject, owner isc.AgentID) {
	panic("!!! TODO")
	// args, err := ch.CallView(accounts.ViewAccountObjects.Message(&owner))
	// require.NoError(t, err)
	// nftIDs, err := accounts.ViewAccountObjects.DecodeOutput(args)
	// require.NoError(t, err)
	// require.Contains(t, nftIDs, nft.ID)
	// require.Equal(t, nftBack.Issuer, nft.Issuer)
	// require.Equal(t, nftBack.Metadata, nft.Metadata)
	// require.True(t, nftBack.Owner.Equals(owner))
}

func TestAccounts_TransferNFTAllowance(t *testing.T) {
	t.Skip("!!! TODO")
	// env := solo.New(t, &solo.InitOptions{})
	// ch := env.NewChain()
	//
	// issuerWallet, _ := ch.Env.NewKeyPairWithFunds()
	// initialOwnerWallet, initialOwnerAddress := ch.Env.NewKeyPairWithFunds()
	// initialOwnerAgentID := isc.NewAddressAgentID(initialOwnerAddress)
	//
	// nft, err := ch.Env.MintNFTL1(issuerWallet, initialOwnerAddress, []byte("foobar"))
	// require.NoError(t, err)
	//
	// // deposit the NFT to the chain to the initial owner's account
	// _, err = ch.PostRequestSync(
	// 	solo.NewCallParams(accounts.FuncDeposit.Message()).
	// 		WithObject(nft.ID).
	// 		AddBaseTokens(10*isc.Million).
	// 		WithMaxAffordableGasBudget(),
	// 	initialOwnerWallet)
	// require.NoError(t, err)
	//
	// require.True(t, ch.HasL2NFT(initialOwnerAgentID, nft.ID))
	// checkChainNFTData(t, ch, nft, initialOwnerAgentID)
	//
	// // send an off-ledger request to transfer the NFT to the another account
	// finalOwnerWallet, finalOwnerAddress := ch.Env.NewKeyPairWithFunds()
	// finalOwnerAgentID := isc.NewAddressAgentID(finalOwnerAddress)
	//
	// _, err = ch.PostRequestOffLedger(
	// 	solo.NewCallParams(accounts.FuncTransferAllowanceTo.Message(finalOwnerAgentID)).
	// 		WithAllowance(isc.NewEmptyAssets().AddObject(nft.ID)).
	// 		WithMaxAffordableGasBudget(),
	// 	initialOwnerWallet,
	// )
	// require.NoError(t, err)
	//
	// require.True(t, ch.HasL2NFT(finalOwnerAgentID, nft.ID))
	// require.False(t, ch.HasL2NFT(initialOwnerAgentID, nft.ID))
	// checkChainNFTData(t, ch, nft, finalOwnerAgentID)
	//
	// // withdraw to L1
	// _, err = ch.PostRequestSync(
	// 	solo.NewCallParams(accounts.FuncWithdraw.Message()).
	// 		WithAllowance(isc.NewAssets(1*isc.Million).AddObject(nft.ID)).
	// 		AddBaseTokens(10*isc.Million).
	// 		WithMaxAffordableGasBudget(),
	// 	finalOwnerWallet,
	// )
	// require.NoError(t, err)
	//
	// require.False(t, ch.HasL2NFT(finalOwnerAgentID, nft.ID))
	// require.True(t, env.HasL1NFT(finalOwnerAddress, nft.ID))
}

func TestAccounts_DepositNFTWithMinStorageDeposit(t *testing.T) {
	t.Skip("!!! TODO")
	// env := solo.New(t, &solo.InitOptions{Debug: true, PrintStackTrace: true})
	// ch := env.NewChain()
	//
	// issuerWallet, issuerAddress := env.NewKeyPairWithFunds(env.NewSeedFromTestNameAndTimestamp(t.Name())	//
	// nft, err := env.MintNFTL1(issuerWallet, issuerAddress, []byte("foobar"))
	// require.NoError(t, err)
	// req := solo.NewCallParams(accounts.FuncDeposit.Message()).
	// 	WithObject(nft.ID).
	// 	WithMaxAffordableGasBudget()
	// req.AddBaseTokens(ch.EstimateNeededStorageDeposit(req, issuerWallet))
	// _, err = ch.PostRequestSync(req, issuerWallet)
	// require.NoError(t, err)
	//
	// testdbhash.VerifyContractStateHash(env, accounts.Contract, "", t.Name())
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
