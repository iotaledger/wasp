package testcore

import (
	"fmt"
	"math"
	"math/big"
	"slices"
	"strconv"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/origin"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/testutil/testmisc"
	"github.com/iotaledger/wasp/packages/testutil/utxodb"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

const BaseTokensDepositFee = 100

func TestDeposit(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: true})
	sender, _ := env.NewKeyPairWithFunds(env.NewSeedFromIndex(11))
	ch := env.NewChain()

	err := ch.DepositBaseTokensToL2(100_000, sender)
	require.NoError(t, err)

	rec := ch.LastReceipt()
	require.NotNil(t, rec)
	t.Logf("========= receipt: %s", rec)
	t.Logf("========= burn log:\n%s", rec.GasBurnLog)
}

// allowance shouldn't allow you to bypass gas fees.
func TestDepositCheatAllowance(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: false})
	sender, senderAddr := env.NewKeyPairWithFunds(env.NewSeedFromIndex(11))
	senderAgentID := isc.NewAgentID(senderAddr)
	ch := env.NewChain()

	const baseTokensSent = 1 * isc.Million

	// send a request where allowance == assets - so that no base tokens are available outside allowance
	_, err := ch.PostRequestSync(
		solo.NewCallParams(accounts.Contract.Name, accounts.FuncDeposit.Name).
			AddBaseTokens(baseTokensSent).
			WithGasBudget(100_000).
			AddAllowanceBaseTokens(baseTokensSent),
		sender,
	)
	require.Error(t, err)

	rec := ch.LastReceipt()
	finalBalance := ch.L2BaseTokens(senderAgentID)
	require.Less(t, finalBalance, baseTokensSent)
	require.EqualValues(t, baseTokensSent, finalBalance+rec.GasFeeCharged)
}

func TestWithdrawEverything(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: true})
	sender, senderAddr := env.NewKeyPairWithFunds(env.NewSeedFromIndex(11))
	senderAgentID := isc.NewAgentID(senderAddr)
	ch := env.NewChain()

	// deposit some base tokens to L2
	initialL1balance := ch.Env.L1BaseTokens(senderAddr)
	baseTokensToDepositToL2 := uint64(100_000)
	err := ch.DepositBaseTokensToL2(baseTokensToDepositToL2, sender)
	require.NoError(t, err)

	depositGasFee := ch.LastReceipt().GasFeeCharged
	l2balance := ch.L2BaseTokens(senderAgentID)

	// construct request with low allowance (just sufficient for storage deposit balance), so its possible to estimate the gas fees
	req := solo.NewCallParams(accounts.Contract.Name, accounts.FuncWithdraw.Name).
		WithFungibleTokens(isc.NewAssetsBaseTokens(l2balance)).AddAllowance(isc.NewAssetsBaseTokens(5200))

	gasEstimate, fee, err := ch.EstimateGasOffLedger(req, sender, true)
	require.NoError(t, err)

	// set the allowance to the maximum possible value
	req = req.WithAllowance(isc.NewAssetsBaseTokens(l2balance - fee)).
		WithGasBudget(gasEstimate)

	_, err = ch.PostRequestOffLedger(req, sender)
	require.NoError(t, err)

	withdrawalGasFee := ch.LastReceipt().GasFeeCharged
	finalL1Balance := ch.Env.L1BaseTokens(senderAddr)
	finalL2Balance := ch.L2BaseTokens(senderAgentID)

	// ensure everything was withdrawn
	require.Equal(t, initialL1balance, finalL1Balance+depositGasFee+withdrawalGasFee)
	require.Zero(t, finalL2Balance)
}

func TestFoundries(t *testing.T) {
	var env *solo.Solo
	var ch *solo.Chain
	var senderKeyPair *cryptolib.KeyPair
	var senderAddr iotago.Address
	var senderAgentID isc.AgentID

	initTest := func() {
		env = solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: true})
		ch, _ = env.NewChainExt(nil, 10*isc.Million, "chain1")
		senderKeyPair, senderAddr = env.NewKeyPairWithFunds(env.NewSeedFromIndex(10))
		senderAgentID = isc.NewAgentID(senderAddr)

		ch.MustDepositBaseTokensToL2(10*isc.Million, senderKeyPair)
	}
	t.Run("newFoundry fails when no allowance is provided", func(t *testing.T) {
		env = solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: true})
		ch, _ = env.NewChainExt(nil, 100_000, "chain1")

		req := solo.NewCallParams(accounts.Contract.Name, accounts.FuncFoundryCreateNew.Name,
			accounts.ParamTokenScheme, codec.EncodeTokenScheme(
				&iotago.SimpleTokenScheme{MaximumSupply: big.NewInt(1), MintedTokens: util.Big0, MeltedTokens: util.Big0},
			),
		).AddBaseTokens(2 * isc.Million).WithGasBudget(math.MaxUint64)
		_, err := ch.PostRequestSync(req, nil)
		require.Error(t, err)
		// it succeeds when allowance is added
		_, err = ch.PostRequestSync(req.AddAllowanceBaseTokens(1*isc.Million), nil)
		require.NoError(t, err)
	})
	t.Run("newFoundry overrides bad melted/minted token counters in tokenscheme", func(t *testing.T) {
		env = solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: true})
		ch, _ = env.NewChainExt(nil, 100_000, "chain1")

		req := solo.NewCallParams(accounts.Contract.Name, accounts.FuncFoundryCreateNew.Name,
			accounts.ParamTokenScheme, codec.EncodeTokenScheme(
				&iotago.SimpleTokenScheme{MaximumSupply: big.NewInt(1), MintedTokens: big.NewInt(10), MeltedTokens: big.NewInt(10)},
			),
		).AddBaseTokens(2 * isc.Million).WithGasBudget(math.MaxUint64)
		_, err := ch.PostRequestSync(req.AddAllowanceBaseTokens(1*isc.Million), nil)
		require.NoError(t, err)
	})
	t.Run("supply 10", func(t *testing.T) {
		initTest()
		sn, _, err := ch.NewFoundryParams(10).
			WithUser(senderKeyPair).
			CreateFoundry()
		require.NoError(t, err)
		require.EqualValues(t, 1, int(sn))
	})
	t.Run("supply 1", func(t *testing.T) {
		initTest()
		sn, _, err := ch.NewFoundryParams(1).
			WithUser(senderKeyPair).
			CreateFoundry()
		require.NoError(t, err)
		require.EqualValues(t, 1, sn)
	})
	t.Run("supply 0", func(t *testing.T) {
		initTest()
		_, _, err := ch.NewFoundryParams(0).
			WithUser(senderKeyPair).
			CreateFoundry()
		testmisc.RequireErrorToBe(t, err, vm.ErrCreateFoundryMaxSupplyMustBePositive)
	})
	t.Run("supply negative", func(t *testing.T) {
		initTest()
		require.Panics(t, func() {
			_, _, _ = ch.NewFoundryParams(-1).
				WithUser(senderKeyPair).
				CreateFoundry()
		})
	})
	t.Run("supply max possible", func(t *testing.T) {
		initTest()
		sn, _, err := ch.NewFoundryParams(abi.MaxUint256).
			WithUser(senderKeyPair).
			CreateFoundry()
		require.NoError(t, err)
		require.EqualValues(t, 1, sn)
	})
	t.Run("supply exceed max possible", func(t *testing.T) {
		initTest()
		maxSupply := new(big.Int).Set(util.MaxUint256)
		maxSupply.Add(maxSupply, big.NewInt(1))
		require.Panics(t, func() {
			_, _, _ = ch.NewFoundryParams(maxSupply).CreateFoundry()
		})
	})
	t.Run("max supply 10, mintTokens 5", func(t *testing.T) {
		initTest()
		sn, nativeTokenID, err := ch.NewFoundryParams(10).
			WithUser(senderKeyPair).
			CreateFoundry()
		require.NoError(t, err)
		require.EqualValues(t, 1, sn)
		ch.AssertL2NativeTokens(senderAgentID, nativeTokenID, util.Big0)
		ch.AssertL2TotalNativeTokens(nativeTokenID, util.Big0)

		err = ch.SendFromL1ToL2AccountBaseTokens(BaseTokensDepositFee, 1000, accounts.CommonAccount(), senderKeyPair)
		require.NoError(t, err)
		t.Logf("common account base tokens = %d before mint", ch.L2CommonAccountBaseTokens())

		err = ch.MintTokens(sn, big.NewInt(5), senderKeyPair)
		require.NoError(t, err)

		ch.AssertL2NativeTokens(senderAgentID, nativeTokenID, big.NewInt(5))
		ch.AssertL2TotalNativeTokens(nativeTokenID, big.NewInt(5))
	})
	t.Run("max supply 1, mintTokens 1", func(t *testing.T) {
		initTest()
		sn, nativeTokenID, err := ch.NewFoundryParams(1).
			WithUser(senderKeyPair).
			CreateFoundry()
		require.NoError(t, err)
		require.EqualValues(t, 1, sn)
		ch.AssertL2NativeTokens(senderAgentID, nativeTokenID, util.Big0)
		ch.AssertL2TotalNativeTokens(nativeTokenID, util.Big0)

		err = ch.SendFromL1ToL2AccountBaseTokens(BaseTokensDepositFee, 1000, accounts.CommonAccount(), senderKeyPair)
		require.NoError(t, err)
		err = ch.MintTokens(sn, 1, senderKeyPair)
		require.NoError(t, err)

		ch.AssertL2NativeTokens(senderAgentID, nativeTokenID, big.NewInt(1))
		ch.AssertL2TotalNativeTokens(nativeTokenID, big.NewInt(1))
	})

	t.Run("max supply 1, mintTokens 2", func(t *testing.T) {
		initTest()
		sn, nativeTokenID, err := ch.NewFoundryParams(1).
			WithUser(senderKeyPair).
			CreateFoundry()
		require.NoError(t, err)
		require.EqualValues(t, 1, sn)

		err = ch.MintTokens(sn, 2, senderKeyPair)
		testmisc.RequireErrorToBe(t, err, vm.ErrNativeTokenSupplyOutOffBounds)

		ch.AssertL2NativeTokens(senderAgentID, nativeTokenID, util.Big0)
		ch.AssertL2TotalNativeTokens(nativeTokenID, util.Big0)
	})
	t.Run("max supply 1000, mintTokens 500_500_1", func(t *testing.T) {
		initTest()
		sn, nativeTokenID, err := ch.NewFoundryParams(1000).
			WithUser(senderKeyPair).
			CreateFoundry()
		require.NoError(t, err)
		require.EqualValues(t, 1, sn)

		err = ch.SendFromL1ToL2AccountBaseTokens(BaseTokensDepositFee, 1000, accounts.CommonAccount(), senderKeyPair)
		require.NoError(t, err)
		err = ch.MintTokens(sn, 500, senderKeyPair)
		require.NoError(t, err)
		ch.AssertL2NativeTokens(senderAgentID, nativeTokenID, big.NewInt(500))
		ch.AssertL2TotalNativeTokens(nativeTokenID, big.NewInt(500))

		err = ch.MintTokens(sn, 500, senderKeyPair)
		require.NoError(t, err)
		ch.AssertL2NativeTokens(senderAgentID, nativeTokenID, 1000)
		ch.AssertL2TotalNativeTokens(nativeTokenID, 1000)

		err = ch.MintTokens(sn, 1, senderKeyPair)
		testmisc.RequireErrorToBe(t, err, vm.ErrNativeTokenSupplyOutOffBounds)

		ch.AssertL2NativeTokens(senderAgentID, nativeTokenID, 1000)
		ch.AssertL2TotalNativeTokens(nativeTokenID, 1000)
	})
	t.Run("max supply MaxUint256, mintTokens MaxUint256_1", func(t *testing.T) {
		initTest()
		sn, nativeTokenID, err := ch.NewFoundryParams(abi.MaxUint256).
			WithUser(senderKeyPair).
			CreateFoundry()
		require.NoError(t, err)
		require.EqualValues(t, 1, sn)

		err = ch.SendFromL1ToL2AccountBaseTokens(BaseTokensDepositFee, 1000, accounts.CommonAccount(), senderKeyPair)
		require.NoError(t, err)
		err = ch.MintTokens(sn, abi.MaxUint256, senderKeyPair)
		require.NoError(t, err)
		ch.AssertL2NativeTokens(senderAgentID, nativeTokenID, abi.MaxUint256)

		err = ch.MintTokens(sn, 1, senderKeyPair)
		testmisc.RequireErrorToBe(t, err, vm.ErrOverflow)

		ch.AssertL2NativeTokens(senderAgentID, nativeTokenID, abi.MaxUint256)
		ch.AssertL2TotalNativeTokens(nativeTokenID, abi.MaxUint256)
	})
	t.Run("max supply 100, destroy fail", func(t *testing.T) {
		initTest()
		sn, nativeTokenID, err := ch.NewFoundryParams(abi.MaxUint256).
			WithUser(senderKeyPair).
			CreateFoundry()
		require.NoError(t, err)
		require.EqualValues(t, 1, sn)

		err = ch.DestroyTokensOnL2(nativeTokenID, big.NewInt(1), senderKeyPair)
		testmisc.RequireErrorToBe(t, err, accounts.ErrNotEnoughFunds)
		ch.AssertL2NativeTokens(senderAgentID, nativeTokenID, util.Big0)
		ch.AssertL2TotalNativeTokens(nativeTokenID, util.Big0)
	})
	t.Run("max supply 100, mint_20, destroy_10", func(t *testing.T) {
		initTest()
		sn, nativeTokenID, err := ch.NewFoundryParams(100).
			WithUser(senderKeyPair).
			CreateFoundry()
		require.NoError(t, err)
		require.EqualValues(t, 1, sn)

		out, err := ch.GetFoundryOutput(1)
		require.NoError(t, err)
		require.EqualValues(t, out.MustNativeTokenID(), nativeTokenID)
		ch.AssertL2NativeTokens(senderAgentID, nativeTokenID, util.Big0)
		ch.AssertL2TotalNativeTokens(nativeTokenID, util.Big0)

		err = ch.SendFromL1ToL2AccountBaseTokens(BaseTokensDepositFee, 1000, accounts.CommonAccount(), senderKeyPair)
		require.NoError(t, err)
		err = ch.MintTokens(sn, 20, senderKeyPair)
		require.NoError(t, err)
		ch.AssertL2NativeTokens(senderAgentID, nativeTokenID, 20)
		ch.AssertL2TotalNativeTokens(nativeTokenID, 20)

		err = ch.DestroyTokensOnL2(nativeTokenID, 10, senderKeyPair)
		require.NoError(t, err)
		ch.AssertL2TotalNativeTokens(nativeTokenID, 10)
		ch.AssertL2NativeTokens(senderAgentID, nativeTokenID, 10)
	})
	t.Run("max supply 1000000, mint_1000000, destroy_1000000", func(t *testing.T) {
		initTest()
		sn, nativeTokenID, err := ch.NewFoundryParams(1_000_000).
			WithUser(senderKeyPair).
			CreateFoundry()
		require.NoError(t, err)
		require.EqualValues(t, 1, sn)

		out, err := ch.GetFoundryOutput(1)
		require.NoError(t, err)
		require.EqualValues(t, out.MustNativeTokenID(), nativeTokenID)
		ch.AssertL2NativeTokens(senderAgentID, nativeTokenID, 0)
		ch.AssertL2TotalNativeTokens(nativeTokenID, 0)

		err = ch.SendFromL1ToL2AccountBaseTokens(BaseTokensDepositFee, 1000, accounts.CommonAccount(), senderKeyPair)
		require.NoError(t, err)
		err = ch.MintTokens(sn, 1_000_000, senderKeyPair)
		require.NoError(t, err)
		ch.AssertL2NativeTokens(senderAgentID, nativeTokenID, big.NewInt(1_000_000))
		ch.AssertL2TotalNativeTokens(nativeTokenID, big.NewInt(1_000_000))
		out, err = ch.GetFoundryOutput(1)
		require.NoError(t, err)
		ts := util.MustTokenScheme(out.TokenScheme)
		require.True(t, big.NewInt(1_000_000).Cmp(ts.MintedTokens) == 0)

		// FIXME bug iotago can't destroy foundry
		// err = destroyTokens(sn, big.NewInt(1000000))
		// require.NoError(t, err)
		// ch.AssertL2TotalNativeTokens(nativeTokenID, util.Big0)
		// ch.AssertL2NativeTokens(userAgentID, nativeTokenID, util.Big0)
		// out, err = ch.GetFoundryOutput(1)
		// require.NoError(t, err)
		// require.True(t, util.Big0.Cmp(out.MintedTokens) == 0)
	})
	t.Run("10 foundries", func(t *testing.T) {
		initTest()
		ch.MustDepositBaseTokensToL2(50_000_000, senderKeyPair)
		nativeTokenIDs := make([]iotago.NativeTokenID, 11)
		for sn := uint32(1); sn <= 10; sn++ {
			snBack, nativeTokenID, err := ch.NewFoundryParams(uint64(sn + 1)).
				WithUser(senderKeyPair).
				CreateFoundry()
			nativeTokenIDs[sn] = nativeTokenID
			require.NoError(t, err)
			require.EqualValues(t, int(sn), int(snBack))
			ch.AssertL2NativeTokens(senderAgentID, nativeTokenID, util.Big0)
			ch.AssertL2TotalNativeTokens(nativeTokenID, util.Big0)
		}
		// mint max supply from each
		ch.MustDepositBaseTokensToL2(50_000_000, senderKeyPair)
		for sn := uint32(1); sn <= 10; sn++ {
			err := ch.MintTokens(sn, sn+1, senderKeyPair)
			require.NoError(t, err)

			out, err := ch.GetFoundryOutput(sn)
			require.NoError(t, err)

			require.EqualValues(t, sn, out.SerialNumber)
			ts := util.MustTokenScheme(out.TokenScheme)
			require.True(t, ts.MaximumSupply.Cmp(big.NewInt(int64(sn+1))) == 0)
			require.True(t, ts.MintedTokens.Cmp(big.NewInt(int64(sn+1))) == 0)
			nativeTokenID := out.MustNativeTokenID()

			ch.AssertL2NativeTokens(senderAgentID, nativeTokenID, big.NewInt(int64(sn+1)))
			ch.AssertL2TotalNativeTokens(nativeTokenID, big.NewInt(int64(sn+1)))
		}
		// destroy 1 token of each nativeTokenID
		for sn := uint32(1); sn <= 10; sn++ {
			err := ch.DestroyTokensOnL2(nativeTokenIDs[sn], big.NewInt(1), senderKeyPair)
			require.NoError(t, err)
		}
		// check balances
		for sn := uint32(1); sn <= 10; sn++ {
			out, err := ch.GetFoundryOutput(sn)
			require.NoError(t, err)

			require.EqualValues(t, sn, out.SerialNumber)
			ts := util.MustTokenScheme(out.TokenScheme)
			require.True(t, ts.MaximumSupply.Cmp(big.NewInt(int64(sn+1))) == 0)
			require.True(t, big.NewInt(0).Sub(ts.MintedTokens, ts.MeltedTokens).Cmp(big.NewInt(int64(sn))) == 0)
			nativeTokenID := out.MustNativeTokenID()

			ch.AssertL2NativeTokens(senderAgentID, nativeTokenID, big.NewInt(int64(sn)))
			ch.AssertL2TotalNativeTokens(nativeTokenID, big.NewInt(int64(sn)))
		}
	})
	t.Run("constant storage deposit to hold a token UTXO", func(t *testing.T) {
		initTest()
		// create a foundry for the maximum amount of tokens possible
		sn, nativeTokenID, err := ch.NewFoundryParams(util.MaxUint256).
			WithUser(senderKeyPair).
			CreateFoundry()
		require.NoError(t, err)

		err = ch.SendFromL1ToL2AccountBaseTokens(BaseTokensDepositFee, 1, accounts.CommonAccount(), senderKeyPair)
		require.NoError(t, err)
		x := ch.L2CommonAccountBaseTokens()
		t.Logf("common account base tokens = %d before mint", x)

		big1 := big.NewInt(1)
		err = ch.MintTokens(sn, big1, senderKeyPair)
		require.NoError(t, err)

		ch.AssertL2NativeTokens(senderAgentID, nativeTokenID, big1)
		ch.AssertL2TotalNativeTokens(nativeTokenID, big1)
		ownerBal1 := ch.L2Assets(ch.OriginatorAgentID)
		commonAccountBalanceBeforeLastMint := ch.L2CommonAccountBaseTokens()

		// after minting 1 token, try to mint the remaining tokens
		allOtherTokens := new(big.Int).Set(util.MaxUint256)
		allOtherTokens = allOtherTokens.Sub(allOtherTokens, big1)

		err = ch.MintTokens(sn, allOtherTokens, senderKeyPair)
		require.NoError(t, err)

		commonAccountBalanceAfterLastMint := ch.L2CommonAccountBaseTokens()
		require.Equal(t, commonAccountBalanceAfterLastMint, commonAccountBalanceBeforeLastMint)
		// assert that no extra base tokens were used for the storage deposit
		ownerBal2 := ch.L2Assets(ch.OriginatorAgentID)
		receipt := ch.LastReceipt()
		require.Equal(t, ownerBal1.BaseTokens+receipt.GasFeeCharged, ownerBal2.BaseTokens)
	})
	t.Run("newFoundry exposes foundry serial number in event", func(t *testing.T) {
		initTest()
		sn, _, err := ch.NewFoundryParams(abi.MaxUint256).
			WithUser(senderKeyPair).
			CreateFoundry()
		require.NoError(t, err)
		require.EqualValues(t, 1, sn)

		events, err := ch.GetEventsForContract(accounts.Contract.Name)
		require.NoError(t, err)
		require.Len(t, events, 1)
		sn, err = codec.DecodeUint32(events[0].Payload)
		require.NoError(t, err)
		require.EqualValues(t, 1, sn)
	})
}

func TestAccountBalances(t *testing.T) {
	env := solo.New(t)

	chainOwner, chainOwnerAddr := env.NewKeyPairWithFunds(env.NewSeedFromIndex(10))
	chainOwnerAgentID := isc.NewAgentID(chainOwnerAddr)

	sender, senderAddr := env.NewKeyPairWithFunds(env.NewSeedFromIndex(11))
	senderAgentID := isc.NewAgentID(senderAddr)

	l1BaseTokens := func(addr iotago.Address) uint64 { return env.L1Assets(addr).BaseTokens }
	totalBaseTokens := l1BaseTokens(chainOwnerAddr) + l1BaseTokens(senderAddr)

	ch, _ := env.NewChainExt(chainOwner, 0, "chain1")

	totalGasFeeCharged := uint64(0)

	checkBalance := func() {
		require.EqualValues(t,
			totalBaseTokens,
			l1BaseTokens(chainOwnerAddr)+l1BaseTokens(senderAddr)+l1BaseTokens(ch.ChainID.AsAddress()),
		)

		anchor := ch.GetAnchorOutputFromL1().GetAliasOutput()
		require.EqualValues(t, l1BaseTokens(ch.ChainID.AsAddress()), anchor.Deposit())

		require.LessOrEqual(t, len(ch.L2Accounts()), 3)

		bi := ch.GetLatestBlockInfo()

		anchorSD := parameters.L1().Protocol.RentStructure.MinRent(anchor)
		require.EqualValues(t,
			anchor.Deposit(),
			anchorSD+ch.L2BaseTokens(chainOwnerAgentID)+ch.L2BaseTokens(senderAgentID)+ch.L2BaseTokens(accounts.CommonAccount()),
		)

		totalGasFeeCharged += bi.GasFeeCharged
		require.EqualValues(t,
			utxodb.FundsFromFaucetAmount+totalGasFeeCharged-anchorSD,
			l1BaseTokens(chainOwnerAddr)+ch.L2BaseTokens(chainOwnerAgentID)+ch.L2BaseTokens(accounts.CommonAccount()),
		)
		require.EqualValues(t,
			utxodb.FundsFromFaucetAmount-totalGasFeeCharged,
			l1BaseTokens(senderAddr)+ch.L2BaseTokens(senderAgentID),
		)
	}

	// preload sender account with base tokens in order to be able to pay for gas fees
	err := ch.DepositBaseTokensToL2(100_000, sender)
	require.NoError(t, err)

	checkBalance()

	for i := 0; i < 5; i++ {
		blobData := fmt.Sprintf("dummy blob data #%d", i+1)
		_, err := ch.UploadBlob(sender, "field", blobData)
		require.NoError(t, err)

		checkBalance()
	}
}

type testParams struct {
	env               *solo.Solo
	chainOwner        *cryptolib.KeyPair
	chainOwnerAddr    iotago.Address
	chainOwnerAgentID isc.AgentID
	user              *cryptolib.KeyPair
	userAddr          iotago.Address
	userAgentID       isc.AgentID
	ch                *solo.Chain
	req               *solo.CallParams
	sn                uint32
	nativeTokenID     iotago.NativeTokenID
}

func initDepositTest(t *testing.T, originParams dict.Dict, initLoad ...uint64) *testParams {
	ret := &testParams{}
	ret.env = solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: true, Debug: true, PrintStackTrace: true})

	ret.chainOwner, ret.chainOwnerAddr = ret.env.NewKeyPairWithFunds(ret.env.NewSeedFromIndex(10))
	ret.chainOwnerAgentID = isc.NewAgentID(ret.chainOwnerAddr)
	ret.user, ret.userAddr = ret.env.NewKeyPairWithFunds(ret.env.NewSeedFromIndex(11))
	ret.userAgentID = isc.NewAgentID(ret.userAddr)

	initBaseTokens := uint64(0)
	if len(initLoad) != 0 {
		initBaseTokens = initLoad[0]
	}
	ret.ch, _ = ret.env.NewChainExt(ret.chainOwner, initBaseTokens, "chain1", originParams)

	ret.req = solo.NewCallParams(accounts.Contract.Name, accounts.FuncDeposit.Name)
	return ret
}

func (v *testParams) createFoundryAndMint(maxSupply, amount interface{}) (uint32, iotago.NativeTokenID) {
	sn, nativeTokenID, err := v.ch.NewFoundryParams(maxSupply).
		WithUser(v.user).
		CreateFoundry()
	require.NoError(v.env.T, err)
	// mint some tokens for the user
	err = v.ch.MintTokens(sn, amount, v.user)
	require.NoError(v.env.T, err)
	// check the balance of the user
	v.ch.AssertL2NativeTokens(v.userAgentID, nativeTokenID, amount)
	require.True(v.env.T, v.ch.L2BaseTokens(v.userAgentID) > 100) // must be some coming from storage deposits
	return sn, nativeTokenID
}

func TestDepositBaseTokens(t *testing.T) {
	// the test check how request transaction construction functions adjust base tokens to the minimum needed for the
	// storage deposit. If storage deposit is 185, anything below that fill be topped up to 185, above that no adjustment is needed
	for _, addBaseTokens := range []uint64{0, 50, 150, 200, 1000} {
		t.Run("add base tokens "+strconv.Itoa(int(addBaseTokens)), func(t *testing.T) {
			v := initDepositTest(t, nil)
			v.req.WithGasBudget(100_000)
			estimatedGas, _, err := v.ch.EstimateGasOnLedger(v.req, v.user)
			require.NoError(t, err)

			v.req.WithGasBudget(estimatedGas)

			v.req = v.req.AddBaseTokens(addBaseTokens)
			tx, _, err := v.ch.PostRequestSyncTx(v.req, v.user)
			require.NoError(t, err)
			rec := v.ch.LastReceipt()

			storageDeposit := parameters.L1().Protocol.RentStructure.MinRent(tx.Essence.Outputs[0])
			t.Logf("byteCost = %d", storageDeposit)

			adjusted := addBaseTokens
			if adjusted < storageDeposit {
				adjusted = storageDeposit
			}
			require.True(t, rec.GasFeeCharged <= adjusted)
			v.ch.AssertL2BaseTokens(v.userAgentID, adjusted-rec.GasFeeCharged)
		})
	}
}

// initWithdrawTest creates foundry with 1_000_000 of max supply and mint 100 tokens to user's account
func initWithdrawTest(t *testing.T, initLoad ...uint64) *testParams {
	v := initDepositTest(t, nil, initLoad...)
	v.ch.MustDepositBaseTokensToL2(2*isc.Million, v.user)
	// create foundry and mint 100 tokens
	v.sn, v.nativeTokenID = v.createFoundryAndMint(1_000_000, 100)
	// prepare request parameters to withdraw everything what is in the account
	// do not run the request yet
	v.req = solo.NewCallParams("accounts", "withdraw").
		AddBaseTokens(12000).
		WithGasBudget(100_000)
	v.printBalances("BEGIN")
	return v
}

func (v *testParams) printBalances(prefix string) {
	v.env.T.Logf("%s: user L1 base tokens: %d", prefix, v.env.L1BaseTokens(v.userAddr))
	v.env.T.Logf("%s: user L1 tokens: %s : %d", prefix, v.nativeTokenID, v.env.L1NativeTokens(v.userAddr, v.nativeTokenID))
	v.env.T.Logf("%s: user L2: %s", prefix, v.ch.L2Assets(v.userAgentID))
	v.env.T.Logf("%s: common account L2: %s", prefix, v.ch.L2CommonAccountAssets())
}

func TestWithdrawDepositNativeTokens(t *testing.T) {
	t.Run("withdraw with empty", func(t *testing.T) {
		v := initWithdrawTest(t, 2*isc.Million)
		_, err := v.ch.PostRequestSync(v.req, v.user)
		testmisc.RequireErrorToBe(t, err, "not enough allowance")
	})
	t.Run("withdraw not enough for storage deposit", func(t *testing.T) {
		v := initWithdrawTest(t, 2*isc.Million)
		v.req.AddAllowanceNativeTokensVect(&iotago.NativeToken{
			ID:     v.nativeTokenID,
			Amount: new(big.Int).SetUint64(10),
		})
		_, err := v.ch.PostRequestSync(v.req, v.user)
		testmisc.RequireErrorToBe(t, err, accounts.ErrNotEnoughBaseTokensForStorageDeposit)
	})
	t.Run("withdraw almost all", func(t *testing.T) {
		v := initWithdrawTest(t, 2*isc.Million)
		// we want to withdraw as many base tokens as possible, so we add 300 because some more will come
		// with assets attached to the 'withdraw' request. However, withdraw all is not possible due to gas
		toWithdraw := v.ch.L2Assets(v.userAgentID).AddBaseTokens(200)
		t.Logf("assets to withdraw: %s", toWithdraw.String())
		// withdraw all tokens to L1, but we do not add base tokens to allowance, so not enough for storage deposit
		v.req.AddAllowance(toWithdraw)
		v.req.AddBaseTokens(BaseTokensDepositFee)
		_, err := v.ch.PostRequestSync(v.req, v.user)
		require.NoError(t, err)
		v.printBalances("END")
	})
	t.Run("mint withdraw destroy fail", func(t *testing.T) {
		v := initWithdrawTest(t, 2*isc.Million)
		allSenderAssets := v.ch.L2Assets(v.userAgentID)
		v.req.AddAllowance(allSenderAssets)
		v.req.AddBaseTokens(BaseTokensDepositFee)
		_, err := v.ch.PostRequestSync(v.req, v.user)
		require.NoError(t, err)

		v.printBalances("AFTER MINT")
		v.env.AssertL1NativeTokens(v.userAddr, v.nativeTokenID, 100)

		// should fail because those tokens are not on the user's on chain account
		err = v.ch.DestroyTokensOnL2(v.nativeTokenID, big.NewInt(50), v.user)
		testmisc.RequireErrorToBe(t, err, accounts.ErrNotEnoughFunds)
		v.env.AssertL1NativeTokens(v.userAddr, v.nativeTokenID, big.NewInt(100))
		v.printBalances("AFTER DESTROY")
	})
	t.Run("mint withdraw destroy success 1", func(t *testing.T) {
		v := initWithdrawTest(t, 2*isc.Million)

		allSenderAssets := v.ch.L2Assets(v.userAgentID)
		v.req.AddAllowance(allSenderAssets)
		v.req.AddBaseTokens(BaseTokensDepositFee)
		_, err := v.ch.PostRequestSync(v.req, v.user)
		require.NoError(t, err)
		v.printBalances("AFTER MINT")
		v.env.AssertL1NativeTokens(v.userAddr, v.nativeTokenID, 100)
		v.ch.AssertL2NativeTokens(v.userAgentID, v.nativeTokenID, 0)

		err = v.ch.DepositAssetsToL2(isc.NewEmptyAssets().AddNativeTokens(v.nativeTokenID, 50), v.user)
		require.NoError(t, err)
		v.env.AssertL1NativeTokens(v.userAddr, v.nativeTokenID, 50)
		v.ch.AssertL2NativeTokens(v.userAgentID, v.nativeTokenID, 50)
		v.ch.AssertL2TotalNativeTokens(v.nativeTokenID, 50)
		v.printBalances("AFTER DEPOSIT")

		err = v.ch.DestroyTokensOnL2(v.nativeTokenID, 49, v.user)
		require.NoError(t, err)
		v.ch.AssertL2NativeTokens(v.userAgentID, v.nativeTokenID, 1)
		v.env.AssertL1NativeTokens(v.userAddr, v.nativeTokenID, 50)
		v.printBalances("AFTER DESTROY")

		// sent the last 50 tokens to an evm account
		_, someEthereumAddr := solo.NewEthereumAccount()
		someEthereumAgentID := isc.NewEthereumAddressAgentID(v.ch.ChainID, someEthereumAddr)

		err = v.ch.TransferAllowanceTo(isc.NewEmptyAssets().AddNativeTokens(v.nativeTokenID, 50),
			someEthereumAgentID,
			v.user,
		)
		require.NoError(t, err)
		v.ch.AssertL2NativeTokens(v.userAgentID, v.nativeTokenID, 1)
		v.env.AssertL1NativeTokens(v.userAddr, v.nativeTokenID, 0)
		v.ch.AssertL2NativeTokens(someEthereumAgentID, v.nativeTokenID, 50)
	})
	t.Run("unwrap use case", func(t *testing.T) {
		v := initWithdrawTest(t, 2*isc.Million)
		allSenderAssets := v.ch.L2Assets(v.userAgentID)
		v.req.AddAllowance(allSenderAssets)
		v.req.AddBaseTokens(BaseTokensDepositFee)
		_, err := v.ch.PostRequestSync(v.req, v.user)
		require.NoError(t, err)
		v.printBalances("AFTER MINT")
		v.env.AssertL1NativeTokens(v.userAddr, v.nativeTokenID, 100)
		v.ch.AssertL2NativeTokens(v.userAgentID, v.nativeTokenID, 0)

		err = v.ch.DepositAssetsToL2(isc.NewEmptyAssets().AddNativeTokens(v.nativeTokenID, 1), v.user)
		require.NoError(t, err)
		v.printBalances("AFTER DEPOSIT 1")

		err = v.ch.DestroyTokensOnL1(v.nativeTokenID, 49, v.user)
		require.NoError(t, err)
		v.printBalances("AFTER DESTROY")
		v.ch.AssertL2NativeTokens(v.userAgentID, v.nativeTokenID, 1)
		v.env.AssertL1NativeTokens(v.userAddr, v.nativeTokenID, 50)
	})
	t.Run("unwrap use case 2", func(t *testing.T) {
		v := initWithdrawTest(t, 2*isc.Million)
		allSenderAssets := v.ch.L2Assets(v.userAgentID)
		v.req.AddAllowance(allSenderAssets)
		v.req.AddBaseTokens(BaseTokensDepositFee)
		_, err := v.ch.PostRequestSync(v.req, v.user)
		require.NoError(t, err)
		v.printBalances("AFTER MINT")
		// no tokens on chain
		v.env.AssertL1NativeTokens(v.userAddr, v.nativeTokenID, 100)
		v.ch.AssertL2NativeTokens(v.userAgentID, v.nativeTokenID, 0)

		// deposit and destroy on the same req (chain currently doesn't have an internal UTXO for this tokenID)
		err = v.ch.DestroyTokensOnL1(v.nativeTokenID, 49, v.user)
		require.NoError(t, err)
		v.printBalances("AFTER DESTROY")
		v.ch.AssertL2NativeTokens(v.userAgentID, v.nativeTokenID, 0)
		v.env.AssertL1NativeTokens(v.userAddr, v.nativeTokenID, 51)
	})
	t.Run("mint withdraw destroy fail", func(t *testing.T) {
		v := initWithdrawTest(t, 2*isc.Million)
		allSenderAssets := v.ch.L2Assets(v.userAgentID)
		v.req.AddAllowance(allSenderAssets)
		v.req.AddBaseTokens(BaseTokensDepositFee)
		_, err := v.ch.PostRequestSync(v.req, v.user)
		require.NoError(t, err)

		v.printBalances("AFTER MINT")
		v.env.AssertL1NativeTokens(v.userAddr, v.nativeTokenID, 100)
		v.ch.AssertL2NativeTokens(v.userAgentID, v.nativeTokenID, 0)

		err = v.ch.DepositAssetsToL2(isc.NewEmptyAssets().AddNativeTokens(v.nativeTokenID, 50), v.user)
		require.NoError(t, err)
		v.env.AssertL1NativeTokens(v.userAddr, v.nativeTokenID, 50)
		v.ch.AssertL2NativeTokens(v.userAgentID, v.nativeTokenID, 50)
		v.ch.AssertL2TotalNativeTokens(v.nativeTokenID, 50)
		v.printBalances("AFTER DEPOSIT")

		err = v.ch.DestroyTokensOnL2(v.nativeTokenID, 50, v.user)
		require.NoError(t, err)
		v.ch.AssertL2NativeTokens(v.userAgentID, v.nativeTokenID, 0)
		v.env.AssertL1NativeTokens(v.userAddr, v.nativeTokenID, 50)
	})

	t.Run("accounting UTXOs and pruning", func(t *testing.T) {
		// mint 100 tokens from chain 1 and withdraw those to L1
		v := initWithdrawTest(t, 2*isc.Million)
		{
			allSenderAssets := v.ch.L2Assets(v.userAgentID)
			v.req.AddAllowance(allSenderAssets)
			v.req.AddBaseTokens(BaseTokensDepositFee)
			_, err := v.ch.PostRequestSync(v.req, v.user)
			require.NoError(t, err)
			v.env.AssertL1NativeTokens(v.userAddr, v.nativeTokenID, 100)
			v.ch.AssertL2NativeTokens(v.userAgentID, v.nativeTokenID, 0)
		}

		// create a new chain (ch2) with active state pruning set to keep only 1 block
		blockKeepAmount := int32(1)
		ch2, _ := v.env.NewChainExt(nil, 0, "evmchain", dict.Dict{
			origin.ParamBlockKeepAmount: codec.EncodeInt32(blockKeepAmount),
		})

		// deposit 1 native token from L1 into ch2
		err := ch2.DepositAssetsToL2(isc.NewAssets(1*isc.Million, iotago.NativeTokens{
			{ID: v.nativeTokenID, Amount: big.NewInt(1)},
		}), v.user)
		require.NoError(t, err)

		// make the chain produce 2 blocks (prune the previous block with the initial deposit info)
		for i := 0; i < 2; i++ {
			_, err = ch2.PostRequestSync(solo.NewCallParams("contract", "func"), nil)
			require.Error(t, err)                      // dummy request, so an error is expected
			require.NotNil(t, ch2.LastReceipt().Error) // but it produced a receipt, thus make the state progress
		}

		// deposit 1 more after the initial deposit block has been prunned
		err = ch2.DepositAssetsToL2(isc.NewAssets(1*isc.Million, iotago.NativeTokens{
			{ID: v.nativeTokenID, Amount: big.NewInt(1)},
		}), v.user)
		require.NoError(t, err)
	})
}

func TestTransferAndCheckBaseTokens(t *testing.T) {
	// initializes it all and prepares withdraw request, does not post it
	v := initWithdrawTest(t, 10_000)
	initialCommonAccountBaseTokens := v.ch.L2CommonAccountAssets().BaseTokens
	initialOwnerAccountBaseTokens := v.ch.L2Assets(v.chainOwnerAgentID).BaseTokens

	// deposit some base tokens into the common account
	someUserWallet, _ := v.env.NewKeyPairWithFunds()
	err := v.ch.SendFromL1ToL2Account(11*isc.Million, isc.NewAssetsBaseTokens(10*isc.Million), accounts.CommonAccount(), someUserWallet)
	require.NoError(t, err)
	commonAccBaseTokens := initialCommonAccountBaseTokens + 10*isc.Million
	require.EqualValues(t, commonAccBaseTokens, v.ch.L2CommonAccountAssets().BaseTokens)
	require.EqualValues(t, initialOwnerAccountBaseTokens+v.ch.LastReceipt().GasFeeCharged, v.ch.L2Assets(v.chainOwnerAgentID).BaseTokens)
	require.EqualValues(t, commonAccBaseTokens, v.ch.L2CommonAccountAssets().BaseTokens)
}

func TestFoundryDestroy(t *testing.T) {
	t.Run("destroy existing", func(t *testing.T) {
		v := initDepositTest(t, nil)
		v.ch.MustDepositBaseTokensToL2(2*isc.Million, v.user)
		sn, _, err := v.ch.NewFoundryParams(1_000_000).
			WithUser(v.user).
			CreateFoundry()
		require.NoError(t, err)

		err = v.ch.DestroyFoundry(sn, v.user)
		require.NoError(t, err)
		_, err = v.ch.GetFoundryOutput(sn)
		testmisc.RequireErrorToBe(t, err, "not found")
	})
	t.Run("destroy fail", func(t *testing.T) {
		v := initDepositTest(t, nil)
		err := v.ch.DestroyFoundry(2, v.user)
		testmisc.RequireErrorToBe(t, err, "unauthorized")
	})
}

func TestTransferPartialAssets(t *testing.T) {
	v := initDepositTest(t, nil)
	v.ch.MustDepositBaseTokensToL2(10*isc.Million, v.user)
	// setup a chain with some base tokens and native tokens for user1
	sn, nativeTokenID, err := v.ch.NewFoundryParams(10).
		WithUser(v.user).
		CreateFoundry()
	require.NoError(t, err)
	require.EqualValues(t, 1, int(sn))

	// deposit base tokens for the chain owner (needed for L1 storage deposit to mint tokens)
	err = v.ch.SendFromL1ToL2AccountBaseTokens(BaseTokensDepositFee, 1*isc.Million, accounts.CommonAccount(), v.chainOwner)
	require.NoError(t, err)
	err = v.ch.SendFromL1ToL2AccountBaseTokens(BaseTokensDepositFee, 1*isc.Million, v.userAgentID, v.user)
	require.NoError(t, err)

	err = v.ch.MintTokens(sn, big.NewInt(10), v.user)
	require.NoError(t, err)

	v.ch.AssertL2NativeTokens(v.userAgentID, nativeTokenID, big.NewInt(10))
	v.ch.AssertL2TotalNativeTokens(nativeTokenID, big.NewInt(10))

	// send funds to user2
	user2, user2Addr := v.env.NewKeyPairWithFunds(v.env.NewSeedFromIndex(100))
	user2AgentID := isc.NewAgentID(user2Addr)

	// deposit 1 base token to "create account" for user2 // TODO maybe remove if account creation is not needed
	v.ch.AssertL2BaseTokens(user2AgentID, 0)
	const baseTokensToSend = 3 * isc.Million
	err = v.ch.SendFromL1ToL2AccountBaseTokens(BaseTokensDepositFee, baseTokensToSend, user2AgentID, user2)
	rec := v.ch.LastReceipt()
	require.NoError(t, err)
	v.env.T.Logf("gas fee charged: %d", rec.GasFeeCharged)
	expectedUser2 := BaseTokensDepositFee + baseTokensToSend - rec.GasFeeCharged
	v.ch.AssertL2BaseTokens(user2AgentID, expectedUser2)
	// -----------------------------
	err = v.ch.SendFromL2ToL2Account(
		isc.NewAssets(
			baseTokensToSend,
			iotago.NativeTokens{
				&iotago.NativeToken{
					ID:     nativeTokenID,
					Amount: big.NewInt(9),
				},
			},
		),
		user2AgentID,
		v.user,
	)
	require.NoError(t, err)

	// assert that balances are correct
	v.ch.AssertL2NativeTokens(v.userAgentID, nativeTokenID, big.NewInt(1))
	v.ch.AssertL2NativeTokens(user2AgentID, nativeTokenID, big.NewInt(9))
	v.ch.AssertL2BaseTokens(user2AgentID, expectedUser2+baseTokensToSend)
	v.ch.AssertL2TotalNativeTokens(nativeTokenID, big.NewInt(10))
}

// TestMintedTokensBurn belongs to iota.go
func TestMintedTokensBurn(t *testing.T) {
	const OneMi = 1_000_000

	_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
	aliasIdent1 := tpkg.RandAliasAddress()

	inputIDs := tpkg.RandOutputIDs(3)
	inputs := iotago.OutputSet{
		inputIDs[0]: &iotago.BasicOutput{
			Amount: OneMi,
			Conditions: iotago.UnlockConditions{
				&iotago.AddressUnlockCondition{Address: ident1},
			},
		},
		inputIDs[1]: &iotago.AliasOutput{
			Amount:         OneMi,
			NativeTokens:   nil,
			AliasID:        aliasIdent1.AliasID(),
			StateIndex:     1,
			StateMetadata:  nil,
			FoundryCounter: 1,
			Conditions: iotago.UnlockConditions{
				&iotago.StateControllerAddressUnlockCondition{Address: ident1},
				&iotago.GovernorAddressUnlockCondition{Address: ident1},
			},
			Features: nil,
		},
		inputIDs[2]: &iotago.FoundryOutput{
			Amount:       OneMi,
			NativeTokens: nil,
			SerialNumber: 1,
			TokenScheme: &iotago.SimpleTokenScheme{
				MintedTokens:  big.NewInt(50),
				MeltedTokens:  util.Big0,
				MaximumSupply: big.NewInt(50),
			},
			Conditions: iotago.UnlockConditions{
				&iotago.ImmutableAliasUnlockCondition{Address: aliasIdent1},
			},
			Features: nil,
		},
	}

	// set input BasicOutput NativeToken to 50 which get burned
	foundryNativeTokenID := inputs[inputIDs[2]].(*iotago.FoundryOutput).MustNativeTokenID()
	inputs[inputIDs[0]].(*iotago.BasicOutput).NativeTokens = iotago.NativeTokens{
		{
			ID:     foundryNativeTokenID,
			Amount: new(big.Int).SetInt64(50),
		},
	}

	essence := &iotago.TransactionEssence{
		NetworkID: tpkg.TestNetworkID,
		Inputs:    inputIDs.UTXOInputs(),
		Outputs: iotago.Outputs{
			&iotago.AliasOutput{
				Amount:         OneMi,
				NativeTokens:   nil,
				AliasID:        aliasIdent1.AliasID(),
				StateIndex:     2,
				StateMetadata:  nil,
				FoundryCounter: 1,
				Conditions: iotago.UnlockConditions{
					&iotago.StateControllerAddressUnlockCondition{Address: ident1},
					&iotago.GovernorAddressUnlockCondition{Address: ident1},
				},
				Features: nil,
			},
			&iotago.FoundryOutput{
				Amount:       2 * OneMi,
				NativeTokens: nil,
				SerialNumber: 1,
				TokenScheme: &iotago.SimpleTokenScheme{
					// burn supply by -50
					MintedTokens:  big.NewInt(50),
					MeltedTokens:  big.NewInt(50),
					MaximumSupply: big.NewInt(50),
				},
				Conditions: iotago.UnlockConditions{
					&iotago.ImmutableAliasUnlockCondition{Address: aliasIdent1},
				},
				Features: nil,
			},
		},
	}

	sigs, err := essence.Sign(inputIDs.OrderedSet(inputs).MustCommitment(), ident1AddrKeys)
	require.NoError(t, err)

	tx := &iotago.Transaction{
		Essence: essence,
		Unlocks: iotago.Unlocks{
			&iotago.SignatureUnlock{Signature: sigs[0]},
			&iotago.ReferenceUnlock{Reference: 0},
			&iotago.AliasUnlock{Reference: 1},
		},
	}

	require.NoError(t, tx.SemanticallyValidate(&iotago.SemanticValidationContext{
		ExtParas:   nil,
		WorkingSet: nil,
	}, inputs))
}

func TestNFTAccount(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: true})
	ch := env.NewChain()

	issuerWallet, _ := ch.Env.NewKeyPairWithFunds()
	ownerWallet, ownerAddress := ch.Env.NewKeyPairWithFunds()
	ownerBalance := ch.Env.L1BaseTokens(ownerAddress)

	_, nftInfo, err := ch.Env.MintNFTL1(issuerWallet, ownerAddress, []byte("foobar"))
	require.NoError(t, err)
	nftAddress := nftInfo.NFTID.ToAddress()

	// deposit funds on behalf of the NFT
	const baseTokensToSend = 10 * isc.Million
	req := solo.NewCallParams(accounts.Contract.Name, accounts.FuncDeposit.Name).
		AddBaseTokens(baseTokensToSend).
		WithMaxAffordableGasBudget().
		WithSender(nftAddress)

	_, err = ch.PostRequestSync(req, ownerWallet)
	require.NoError(t, err)
	rec := ch.LastReceipt()

	nftAgentID := isc.NewAgentID(nftAddress)
	ch.AssertL2BaseTokens(nftAgentID, baseTokensToSend-rec.GasFeeCharged)
	ch.Env.AssertL1BaseTokens(nftAddress, 0)
	ch.Env.AssertL1BaseTokens(
		ownerAddress,
		ownerBalance+nftInfo.Output.Deposit()-baseTokensToSend,
	)
	require.True(t, ch.Env.HasL1NFT(ownerAddress, &nftInfo.NFTID))

	// withdraw to the NFT on L1
	const baseTokensToWithdrawal = 1 * isc.Million
	wdReq := solo.NewCallParams(accounts.Contract.Name, accounts.FuncWithdraw.Name).
		AddAllowanceBaseTokens(baseTokensToWithdrawal).
		WithMaxAffordableGasBudget()

	// NFT owner on L1 can't move L2 funds owned by the NFT unless the request is sent in behalf of the NFT (NFTID is specified as "Sender")
	_, err = ch.PostRequestSync(wdReq, ownerWallet)
	require.Error(t, err)

	// NFT owner can withdraw funds owned by the NFT on the chain
	_, err = ch.PostRequestSync(wdReq.WithSender(nftAddress), ownerWallet)
	require.NoError(t, err)
	ch.Env.AssertL1BaseTokens(nftAddress, baseTokensToWithdrawal)
}

func checkChainNFTData(t *testing.T, ch *solo.Chain, nft *isc.NFT, owner isc.AgentID) {
	ret, err := ch.CallView(accounts.Contract.Name, accounts.ViewNFTData.Name, dict.Dict{
		accounts.ParamNFTID: nft.ID[:],
	})
	require.NoError(t, err)
	nftBack, err := isc.NFTFromBytes(ret.Get(accounts.ParamNFTData))
	require.NoError(t, err)
	require.Equal(t, nftBack.ID, nft.ID)
	require.Equal(t, nftBack.Issuer, nft.Issuer)
	require.Equal(t, nftBack.Metadata, nft.Metadata)
	require.True(t, nftBack.Owner.Equals(owner))
}

func TestTransferNFTAllowance(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: true})
	ch := env.NewChain()

	issuerWallet, _ := ch.Env.NewKeyPairWithFunds()
	initialOwnerWallet, initialOwnerAddress := ch.Env.NewKeyPairWithFunds()
	initialOwnerAgentID := isc.NewAgentID(initialOwnerAddress)

	nft, _, err := ch.Env.MintNFTL1(issuerWallet, initialOwnerAddress, []byte("foobar"))
	require.NoError(t, err)

	// deposit the NFT to the chain to the initial owner's account
	_, err = ch.PostRequestSync(
		solo.NewCallParams(accounts.Contract.Name, accounts.FuncDeposit.Name).
			WithNFT(nft).
			AddBaseTokens(10*isc.Million).
			WithMaxAffordableGasBudget(),
		initialOwnerWallet)
	require.NoError(t, err)

	require.True(t, ch.HasL2NFT(initialOwnerAgentID, &nft.ID))
	checkChainNFTData(t, ch, nft, initialOwnerAgentID)

	// send an off-ledger request to transfer the NFT to the another account
	finalOwnerWallet, finalOwnerAddress := ch.Env.NewKeyPairWithFunds()
	finalOwnerAgentID := isc.NewAgentID(finalOwnerAddress)

	_, err = ch.PostRequestOffLedger(
		solo.NewCallParams(accounts.Contract.Name, accounts.FuncTransferAllowanceTo.Name, dict.Dict{
			accounts.ParamAgentID: codec.Encode(finalOwnerAgentID),
		}).
			WithAllowance(isc.NewEmptyAssets().AddNFTs(nft.ID)).
			WithMaxAffordableGasBudget(),
		initialOwnerWallet,
	)
	require.NoError(t, err)

	require.True(t, ch.HasL2NFT(finalOwnerAgentID, &nft.ID))
	require.False(t, ch.HasL2NFT(initialOwnerAgentID, &nft.ID))
	checkChainNFTData(t, ch, nft, finalOwnerAgentID)

	// withdraw to L1
	_, err = ch.PostRequestSync(
		solo.NewCallParams(accounts.Contract.Name, accounts.FuncWithdraw.Name).
			WithAllowance(isc.NewAssets(1*isc.Million, nil, nft.ID)).
			AddBaseTokens(10*isc.Million).
			WithMaxAffordableGasBudget(),
		finalOwnerWallet,
	)
	require.NoError(t, err)

	require.False(t, ch.HasL2NFT(finalOwnerAgentID, &nft.ID))
	require.True(t, env.HasL1NFT(finalOwnerAddress, &nft.ID))
	_, err = ch.CallView(accounts.Contract.Name, accounts.ViewNFTData.Name, dict.Dict{
		accounts.ParamNFTID: nft.ID[:],
	})
	require.Error(t, err)
	require.Regexp(t, "NFTID not found", err.Error())
}

func TestDepositNFTWithMinStorageDeposit(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: false, Debug: true, PrintStackTrace: true})
	ch := env.NewChain()

	issuerWallet, issuerAddress := env.NewKeyPairWithFunds()

	nft, _, err := env.MintNFTL1(issuerWallet, issuerAddress, []byte("foobar"))
	require.NoError(t, err)
	req := solo.NewCallParams(accounts.Contract.Name, accounts.FuncDeposit.Name).
		WithNFT(nft).
		WithMaxAffordableGasBudget()
	req.AddBaseTokens(ch.EstimateNeededStorageDeposit(req, issuerWallet))
	_, err = ch.PostRequestSync(req, issuerWallet)
	require.NoError(t, err)
}

func TestUnprocessableWithNoPruning(t *testing.T) {
	testUnprocessable(t, nil)
}

func TestUnprocessableWithPruning(t *testing.T) {
	testUnprocessable(t, dict.Dict{
		origin.ParamBlockKeepAmount: codec.EncodeInt32(1),
	})
}

func testUnprocessable(t *testing.T, originParams dict.Dict) {
	v := initDepositTest(t, originParams)
	v.ch.MustDepositBaseTokensToL2(2*isc.Million, v.user)
	// create many foundries and mint 1 token on each
	_, nativeTokenID1 := v.createFoundryAndMint(1, 1)
	_, nativeTokenID2 := v.createFoundryAndMint(1, 1)
	_, nativeTokenID3 := v.createFoundryAndMint(1, 1)
	_, nativeTokenID4 := v.createFoundryAndMint(1, 1)

	assets := isc.NewAssets(1*isc.Million, iotago.NativeTokens{
		{ID: nativeTokenID1, Amount: big.NewInt(1)},
		{ID: nativeTokenID2, Amount: big.NewInt(1)},
		{ID: nativeTokenID3, Amount: big.NewInt(1)},
		{ID: nativeTokenID4, Amount: big.NewInt(1)},
	})

	withdrawReq := solo.NewCallParams("accounts", "withdraw").
		WithAllowance(assets).
		WithMaxAffordableGasBudget()
	_, err := v.ch.PostRequestOffLedger(withdrawReq, v.user)
	require.NoError(t, err)

	// ---

	// move the native tokens to a new user that doesn't have on-chain balance
	newUser, newUserAddress := v.env.NewKeyPairWithFunds()
	newUserAgentID := isc.NewAgentID(newUserAddress)
	v.env.SendL1(newUserAddress, assets, v.user)
	// also create an NFT
	iscNFT, _, err := v.ch.Env.MintNFTL1(v.user, newUserAddress, []byte("foobar"))
	require.NoError(t, err)

	newuserL1NativeTokens := v.env.L1Assets(newUserAddress).NativeTokens
	assetsContain := func(tokens iotago.NativeTokens, nativeTokenID iotago.NativeTokenID) bool {
		return lo.ContainsBy(tokens, func(nt *iotago.NativeToken) bool {
			return nt.ID == nativeTokenID
		})
	}
	require.True(t, assetsContain(newuserL1NativeTokens, nativeTokenID1))
	require.True(t, assetsContain(newuserL1NativeTokens, nativeTokenID2))
	require.True(t, assetsContain(newuserL1NativeTokens, nativeTokenID3))
	require.True(t, assetsContain(newuserL1NativeTokens, nativeTokenID4))

	// try to deposit all native tokens in a request with just the minimum SD
	unprocessableReq := solo.NewCallParams(accounts.Contract.Name, accounts.FuncDeposit.Name).
		WithFungibleTokens(isc.NewAssets(0, assets.NativeTokens)).
		WithNFT(iscNFT)

	tx, receipt, _, err := v.ch.PostRequestSyncExt(unprocessableReq, newUser)
	require.Error(t, err)
	testmisc.RequireErrorToBe(t, err, "request has been skipped")
	require.Nil(t, receipt) // nil receipt means the request was not processed

	txReqs, err := v.ch.Env.RequestsForChain(tx, v.ch.ChainID)
	require.NoError(t, err)
	unprocessableReqID := txReqs[0].ID()

	isInUnprocessableList := func() bool {
		res, err2 := v.ch.CallView(
			blocklog.Contract.Name, blocklog.ViewHasUnprocessable.Name,
			blocklog.ParamRequestID, unprocessableReqID,
		)
		require.NoError(t, err2)
		return codec.MustDecodeBool(res.Get(blocklog.ParamUnprocessableRequestExists))
	}

	require.True(t, isInUnprocessableList())

	// assert trying to "retry" the request won't work (still not enough funds)
	retryReq := solo.NewCallParams(
		blocklog.Contract.Name, blocklog.FuncRetryUnprocessable.Name,
		blocklog.ParamRequestID, unprocessableReqID,
	)
	_, rec, _, err := v.ch.PostRequestSyncExt(retryReq, newUser)
	require.NoError(t, err)
	require.Nil(t, rec.Error)
	// the "retry request" is successful, but "there request to be retried" did not produce a receipt, meaning it was skipped again
	// check that the "request to be retried" did not succeed (no receipt, still in the unprocessed list)
	receipt, err = v.ch.GetRequestReceipt(unprocessableReqID)
	require.NoError(t, err)
	require.Nil(t, receipt)
	require.True(t, isInUnprocessableList())
	require.False(t, blocklog.HasUnprocessableRequestBeenRemovedInBlock(v.ch.LatestBlock(), unprocessableReqID)) // assert this function returns false, its used to prevent these requests from being re-added to the mempool on a reorg

	// --
	// deposit funds and retry that request
	err = v.ch.DepositBaseTokensToL2(10*isc.Million, newUser)
	require.NoError(t, err)
	_, rec, _, err = v.ch.PostRequestSyncExt(retryReq, newUser)
	require.NoError(t, err)
	require.Nil(t, rec.Error) // assert the receipt for the "retry req" exists and its successful
	require.Zero(t, rec.SDCharged)

	receipt, err = v.ch.GetRequestReceipt(unprocessableReqID)
	require.NoError(t, err)
	require.NotNil(t, receipt)
	require.Nil(t, receipt.Error)             // assert the receit for the initially unprocessable request exists and is successful
	require.False(t, isInUnprocessableList()) // assert the request was removed from the unprocessable list
	require.NotZero(t, receipt.SDCharged)
	require.True(t, blocklog.HasUnprocessableRequestBeenRemovedInBlock(v.ch.LatestBlock(), unprocessableReqID)) // assert this function returns true, its used to prevent these requests from being re-added to the mempool on a reorg

	// assert the user was credited the tokens from the "initially unprocessable request"
	userAssets := v.ch.L2Assets(newUserAgentID)
	require.Len(t, userAssets.NativeTokens, 4)
	require.True(t, assetsContain(userAssets.NativeTokens, nativeTokenID1))
	require.True(t, assetsContain(userAssets.NativeTokens, nativeTokenID2))
	require.True(t, assetsContain(userAssets.NativeTokens, nativeTokenID3))
	require.True(t, assetsContain(userAssets.NativeTokens, nativeTokenID4))
	require.Len(t, userAssets.NFTs, 1)
	require.EqualValues(t, userAssets.NFTs[0], iscNFT.ID)

	// try the "retry request" again, assert it fails
	_, rec, _, err = v.ch.PostRequestSyncExt(retryReq, newUser)
	require.NoError(t, err)
	require.Error(t, rec.Error)
	require.False(t, blocklog.HasUnprocessableRequestBeenRemovedInBlock(v.ch.LatestBlock(), unprocessableReqID)) // assert this function returns false, its used to prevent these requests from being re-added to the mempool on a reorg

	// --
	// try to withdrawal the native tokens
	err = v.ch.Withdraw(isc.NewAssets(1*isc.Million, userAssets.NativeTokens, iscNFT.ID), newUser)
	require.NoError(t, err)

	require.Len(t, v.ch.L2Assets(newUserAgentID).NativeTokens, 0)
	require.Len(t, v.ch.L2Assets(newUserAgentID).NFTs, 0)
	v.env.AssertL1NativeTokens(newUserAddress, nativeTokenID1, 1)
	v.env.AssertL1NativeTokens(newUserAddress, nativeTokenID2, 1)
	v.env.AssertL1NativeTokens(newUserAddress, nativeTokenID3, 1)
	v.env.AssertL1NativeTokens(newUserAddress, nativeTokenID4, 1)
	require.Len(t, v.env.L1NFTs(newUserAddress), 1)
}

func TestDepositRandomContractMinFee(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: true})
	ch := env.NewChain()

	wallet, addr := ch.Env.NewKeyPairWithFunds()
	agentID := isc.NewAgentID(addr)

	sent := 1 * isc.Million
	_, err := ch.PostRequestSync(solo.NewCallParams("", "").AddBaseTokens(sent), wallet)
	require.Error(t, err)
	receipt := ch.LastReceipt()
	require.Error(t, receipt.Error)

	require.EqualValues(t, gas.DefaultFeePolicy().MinFee(), receipt.GasFeeCharged)
	require.EqualValues(t, sent-receipt.GasFeeCharged, ch.L2BaseTokens(agentID))
}

func TestAllowanceNotEnoughFunds(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: true})
	ch := env.NewChain()

	wallet, _ := ch.Env.NewKeyPairWithFunds()
	allowances := []*isc.Assets{
		// test base token
		isc.NewAssetsBaseTokens(1000 * isc.Million),
		// test fungible tokens
		isc.NewAssets(0, iotago.NativeTokens{&iotago.NativeToken{
			ID:     [38]byte{0x1},
			Amount: big.NewInt(10),
		}}),
		// test NFTs
		isc.NewAssets(0, nil, iotago.NFTID{0x1}),
	}
	for _, a := range allowances {
		_, err := ch.PostRequestSync(
			solo.NewCallParams(accounts.Contract.Name, accounts.FuncDeposit.Name).
				AddBaseTokens(1*isc.Million).
				WithAllowance(a).
				WithMaxAffordableGasBudget(),
			wallet)
		require.Error(t, err)
		testmisc.RequireErrorToBe(t, err, vm.ErrNotEnoughFundsForAllowance)
		receipt := ch.LastReceipt()
		require.EqualValues(t, gas.DefaultFeePolicy().MinFee(), receipt.GasFeeCharged)
	}
}

func TestDepositWithNoGasBudget(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: true})
	senderWallet, _ := env.NewKeyPairWithFunds(env.NewSeedFromIndex(11))
	ch := env.NewChain()

	// try to deposit with 0 gas budget
	_, err := ch.PostRequestSync(
		solo.NewCallParams(accounts.Contract.Name, accounts.FuncDeposit.Name).
			WithFungibleTokens(isc.NewAssetsBaseTokens(2*isc.Million)).
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

func TestRequestWithNoGasBudget(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: true})
	ch := env.NewChain()
	senderWallet, _ := env.NewKeyPairWithFunds()
	req := solo.NewCallParams("dummy", "dummy").WithGasBudget(0)

	// offledger request with 0 gas
	_, err := ch.PostRequestOffLedger(req, senderWallet)
	require.EqualValues(t, 0, ch.LastReceipt().GasBudget)
	testmisc.RequireErrorToBe(t, err, vm.ErrContractNotFound)

	// post the request via on-ledger (the account has funds now), the request gets bumped to "minGasBudget"
	_, err = ch.PostRequestSync(req.WithFungibleTokens(isc.NewAssetsBaseTokens(10*isc.Million)), senderWallet)
	require.EqualValues(t, gas.LimitsDefault.MinGasPerRequest, ch.LastReceipt().GasBudget)
	testmisc.RequireErrorToBe(t, err, vm.ErrContractNotFound)

	// post the request off-ledger again (the account has funds now), the request gets bumped to "minGasBudget"
	_, err = ch.PostRequestOffLedger(req, senderWallet)
	require.EqualValues(t, gas.LimitsDefault.MinGasPerRequest, ch.LastReceipt().GasBudget)
	testmisc.RequireErrorToBe(t, err, vm.ErrContractNotFound)
}

func TestNonces(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: true})
	ch := env.NewChain()
	senderWallet, _ := env.NewKeyPairWithFunds()
	ch.DepositAssetsToL2(isc.NewAssetsBaseTokens(10*isc.Million), senderWallet)

	req := solo.NewCallParams("dummy", "dummy").WithGasBudget(0).WithNonce(0)
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

func TestNFTMint(t *testing.T) {
	env := solo.New(t)
	ch := env.NewChain()

	t.Run("mint for another user", func(t *testing.T) {
		wallet, _ := env.NewKeyPairWithFunds()
		anotherUserAgentID := isc.NewAgentID(tpkg.RandEd25519Address())

		// mint NFT to another user and keep it on chain
		req := solo.NewCallParams(
			accounts.Contract.Name, accounts.FuncMintNFT.Name,
			accounts.ParamNFTImmutableData, []byte("foobar"),
			accounts.ParamAgentID, anotherUserAgentID.Bytes(),
		).
			AddBaseTokens(2 * isc.Million).
			WithAllowance(isc.NewAssetsBaseTokens(1 * isc.Million)).
			WithMaxAffordableGasBudget()

		require.Len(t, ch.L2NFTs(anotherUserAgentID), 0)
		_, err := ch.PostRequestSync(req, wallet)
		require.NoError(t, err)

		// post a dummy request to make the chain progress to the next block
		ch.PostRequestOffLedger(solo.NewCallParams("foo", "bar"), wallet)
		require.Len(t, ch.L2NFTs(anotherUserAgentID), 1)
	})

	t.Run("mint for another user, directly to outside the chain", func(t *testing.T) {
		wallet, _ := env.NewKeyPairWithFunds()

		anotherUserAddr := tpkg.RandEd25519Address()
		anotherUserAgentID := isc.NewAgentID(anotherUserAddr)

		// mint NFT to another user and withdraw it
		req := solo.NewCallParams(
			accounts.Contract.Name, accounts.FuncMintNFT.Name,
			accounts.ParamNFTImmutableData, []byte("foobar"),
			accounts.ParamAgentID, anotherUserAgentID.Bytes(),
			accounts.ParamNFTWithdrawOnMint, codec.Encode(true),
		).
			AddBaseTokens(2 * isc.Million).
			WithAllowance(isc.NewAssetsBaseTokens(1 * isc.Million)).
			WithMaxAffordableGasBudget()

		require.Len(t, env.L1NFTs(anotherUserAddr), 0)
		ret, err := ch.PostRequestSync(req, wallet)
		mintID := ret.Get(accounts.ParamMintID)
		require.NoError(t, err)
		require.Len(t, ch.L2NFTs(anotherUserAgentID), 0)
		userL1NFTs := env.L1NFTs(anotherUserAddr)
		NFTID := iotago.NFTIDFromOutputID(lo.Keys(userL1NFTs)[0])
		require.Len(t, userL1NFTs, 1)

		// post a dummy request to make the chain progress to the next block
		ch.PostRequestOffLedger(solo.NewCallParams("foo", "bar"), wallet)

		// check that the internal ID mapping  matches the L1 NFT
		ret, err = ch.CallView(accounts.Contract.Name, accounts.ViewNFTIDbyMintID.Name,
			accounts.ParamMintID, mintID)
		require.NoError(t, err)
		storedNFTID := ret.Get(accounts.ParamNFTID)
		require.True(t, slices.Equal(storedNFTID, NFTID[:]))
	})

	t.Run("mint to self, then mint from it as a collection", func(t *testing.T) {
		wallet, address := env.NewKeyPairWithFunds()
		agentID := isc.NewAgentID(address)

		// mint NFT to self and keep it on chain
		req := solo.NewCallParams(
			accounts.Contract.Name, accounts.FuncMintNFT.Name,
			accounts.ParamNFTImmutableData, []byte("foobar"),
			accounts.ParamAgentID, agentID.Bytes(),
		).
			AddBaseTokens(2 * isc.Million).
			WithAllowance(isc.NewAssetsBaseTokens(1 * isc.Million)).
			WithMaxAffordableGasBudget()

		require.Len(t, ch.L2NFTs(agentID), 0)
		_, err := ch.PostRequestSync(req, wallet)
		require.NoError(t, err)

		// post a dummy request to make the chain progress to the next block
		ch.PostRequestOffLedger(solo.NewCallParams("foo", "bar"), wallet)
		require.Len(t, env.L1NFTs(address), 0)
		userL2NFTs := ch.L2NFTs(agentID)
		require.Len(t, userL2NFTs, 1)

		// try minting another NFT using the first one as the collection
		fistNFTID := userL2NFTs[0]

		req = solo.NewCallParams(
			accounts.Contract.Name, accounts.FuncMintNFT.Name,
			accounts.ParamNFTImmutableData, []byte("foobar_collection"),
			accounts.ParamAgentID, agentID.Bytes(),
			accounts.ParamCollectionID, codec.Encode(fistNFTID),
		).
			AddBaseTokens(2 * isc.Million).
			WithAllowance(isc.NewAssetsBaseTokens(1 * isc.Million)).
			WithMaxAffordableGasBudget()

		ret, err := ch.PostRequestSync(req, wallet)
		require.NoError(t, err)
		mintID := ret.Get(accounts.ParamMintID)

		// post a dummy request to make the chain progress to the next block
		ch.PostRequestOffLedger(solo.NewCallParams("foo", "bar"), wallet)

		ret, err = ch.CallView(accounts.Contract.Name, accounts.ViewNFTIDbyMintID.Name,
			accounts.ParamMintID, mintID)
		require.NoError(t, err)
		NFTIDInCollection := ret.Get(accounts.ParamNFTID)

		ret, err = ch.CallView(accounts.Contract.Name, accounts.ViewNFTData.Name,
			accounts.ParamNFTID, NFTIDInCollection)
		require.NoError(t, err)

		nftData, err := isc.NFTFromBytes(ret.Get(accounts.ParamNFTData))
		require.NoError(t, err)
		require.True(t, nftData.Issuer.Equal(fistNFTID.ToAddress()))
		require.True(t, nftData.Owner.Equals(agentID))

		// withdraw both NFTs
		err = ch.Withdraw(isc.NewEmptyAssets().AddNFTs(fistNFTID), wallet)
		require.NoError(t, err)

		err = ch.Withdraw(isc.NewEmptyAssets().AddNFTs(iotago.NFTID(NFTIDInCollection)), wallet)
		require.NoError(t, err)

		require.Len(t, env.L1NFTs(address), 2)
		require.Len(t, ch.L2NFTs(agentID), 0)
	})
}
