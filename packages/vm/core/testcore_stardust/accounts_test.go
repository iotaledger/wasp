package testcore

import (
	"fmt"
	"math/big"
	"strconv"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/testutil/testmisc"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/vmcontext/vmtxbuilder"
	"github.com/stretchr/testify/require"
)

func TestFoundries(t *testing.T) {
	var env *solo.Solo
	var ch *solo.Chain
	var senderKeyPair *cryptolib.KeyPair
	var senderAddr iotago.Address
	var senderAgentID *iscp.AgentID

	initTest := func() {
		env = solo.New(t)
		env.EnablePublisher(true)
		ch = env.NewChain(nil, "chain1")
		defer env.WaitPublisher()
		defer ch.Log.Sync()

		senderKeyPair, senderAddr = env.NewKeyPairWithFunds(env.NewSeedFromIndex(10))
		senderAgentID = iscp.NewAgentID(senderAddr, 0)
	}

	t.Run("supply 10", func(t *testing.T) {
		initTest()
		sn, _, err := ch.NewFoundryParams(10).CreateFoundry()
		require.NoError(t, err)
		require.EqualValues(t, 1, int(sn))
	})
	t.Run("supply 1", func(t *testing.T) {
		initTest()
		sn, _, err := ch.NewFoundryParams(1).CreateFoundry()
		require.NoError(t, err)
		require.EqualValues(t, 1, sn)
	})
	t.Run("supply 0", func(t *testing.T) {
		initTest()
		_, _, err := ch.NewFoundryParams(0).CreateFoundry()
		testmisc.RequireErrorToBe(t, err, vmtxbuilder.ErrCreateFoundryMaxSupplyMustBePositive)
	})
	t.Run("supply negative", func(t *testing.T) {
		initTest()
		sn, _, err := ch.NewFoundryParams(-1).CreateFoundry()
		// encoding will ignore sign
		require.NoError(t, err)
		require.EqualValues(t, 1, sn)
	})
	t.Run("supply max possible", func(t *testing.T) {
		initTest()
		sn, _, err := ch.NewFoundryParams(abi.MaxUint256).CreateFoundry()
		require.NoError(t, err)
		require.EqualValues(t, 1, sn)
	})
	t.Run("supply exceed max possible", func(t *testing.T) {
		initTest()
		maxSupply := big.NewInt(0).Set(abi.MaxUint256)
		maxSupply.Add(maxSupply, big.NewInt(1))
		_, _, err := ch.NewFoundryParams(maxSupply).CreateFoundry()
		testmisc.RequireErrorToBe(t, err, vmtxbuilder.ErrCreateFoundryMaxSupplyTooBig)
	})
	// TODO cover all parameter options

	t.Run("max supply 10, mintTokens 5", func(t *testing.T) {
		initTest()
		sn, tokenID, err := ch.NewFoundryParams(10).
			WithUser(senderKeyPair).
			CreateFoundry()
		require.NoError(t, err)
		require.EqualValues(t, 1, sn)
		ch.AssertL2AccountNativeToken(senderAgentID, &tokenID, big.NewInt(0))
		ch.AssertL2TotalNativeTokens(&tokenID, big.NewInt(0))

		err = ch.SendFromL1ToL2AccountIotas(1000, ch.CommonAccount(), senderKeyPair)
		require.NoError(t, err)
		t.Logf("common account iotas = %d before mint", ch.L2CommonAccountIotas())

		err = ch.MintTokens(sn, big.NewInt(5), senderKeyPair)
		require.NoError(t, err)

		ch.AssertL2AccountNativeToken(senderAgentID, &tokenID, big.NewInt(5))
		ch.AssertL2TotalNativeTokens(&tokenID, big.NewInt(5))
	})
	t.Run("max supply 1, mintTokens 1", func(t *testing.T) {
		initTest()
		sn, tokenID, err := ch.NewFoundryParams(1).
			WithUser(senderKeyPair).
			CreateFoundry()
		require.NoError(t, err)
		require.EqualValues(t, 1, sn)
		ch.AssertL2AccountNativeToken(senderAgentID, &tokenID, big.NewInt(0))
		ch.AssertL2TotalNativeTokens(&tokenID, big.NewInt(0))

		err = ch.SendFromL1ToL2AccountIotas(1000, ch.CommonAccount(), senderKeyPair)
		require.NoError(t, err)
		err = ch.MintTokens(sn, 1, senderKeyPair)
		require.NoError(t, err)

		ch.AssertL2AccountNativeToken(senderAgentID, &tokenID, big.NewInt(1))
		ch.AssertL2TotalNativeTokens(&tokenID, big.NewInt(1))
	})

	t.Run("max supply 1, mintTokens 2", func(t *testing.T) {
		initTest()
		sn, tokenID, err := ch.NewFoundryParams(1).
			WithUser(senderKeyPair).
			CreateFoundry()
		require.NoError(t, err)
		require.EqualValues(t, 1, sn)

		err = ch.MintTokens(sn, 2, senderKeyPair)
		testmisc.RequireErrorToBe(t, err, vmtxbuilder.ErrNativeTokenSupplyOutOffBounds)

		ch.AssertL2AccountNativeToken(senderAgentID, &tokenID, big.NewInt(0))
		ch.AssertL2TotalNativeTokens(&tokenID, big.NewInt(0))
	})
	t.Run("max supply 1000, mintTokens 500_500_1", func(t *testing.T) {
		initTest()
		sn, tokenID, err := ch.NewFoundryParams(1000).
			WithUser(senderKeyPair).
			CreateFoundry()
		require.NoError(t, err)
		require.EqualValues(t, 1, sn)

		err = ch.SendFromL1ToL2AccountIotas(1000, ch.CommonAccount(), senderKeyPair)
		require.NoError(t, err)
		err = ch.MintTokens(sn, 500, senderKeyPair)
		require.NoError(t, err)
		ch.AssertL2AccountNativeToken(senderAgentID, &tokenID, big.NewInt(500))
		ch.AssertL2TotalNativeTokens(&tokenID, big.NewInt(500))

		err = ch.MintTokens(sn, 500, senderKeyPair)
		require.NoError(t, err)
		ch.AssertL2AccountNativeToken(senderAgentID, &tokenID, 1000)
		ch.AssertL2TotalNativeTokens(&tokenID, 1000)

		err = ch.MintTokens(sn, 1, senderKeyPair)
		testmisc.RequireErrorToBe(t, err, vmtxbuilder.ErrNativeTokenSupplyOutOffBounds)

		ch.AssertL2AccountNativeToken(senderAgentID, &tokenID, 1000)
		ch.AssertL2TotalNativeTokens(&tokenID, 1000)
	})
	t.Run("max supply MaxUint256, mintTokens MaxUint256_1", func(t *testing.T) {
		initTest()
		sn, tokenID, err := ch.NewFoundryParams(abi.MaxUint256).
			WithUser(senderKeyPair).
			CreateFoundry()
		require.NoError(t, err)
		require.EqualValues(t, 1, sn)

		err = ch.SendFromL1ToL2AccountIotas(1000, ch.CommonAccount(), senderKeyPair)
		require.NoError(t, err)
		err = ch.MintTokens(sn, abi.MaxUint256, senderKeyPair)
		require.NoError(t, err)
		ch.AssertL2AccountNativeToken(senderAgentID, &tokenID, abi.MaxUint256)

		err = ch.MintTokens(sn, 1, senderKeyPair)
		testmisc.RequireErrorToBe(t, err, vmtxbuilder.ErrOverflow)

		ch.AssertL2AccountNativeToken(senderAgentID, &tokenID, abi.MaxUint256)
		ch.AssertL2TotalNativeTokens(&tokenID, abi.MaxUint256)
	})
	t.Run("max supply 100, destroy fail", func(t *testing.T) {
		initTest()
		sn, tokenID, err := ch.NewFoundryParams(abi.MaxUint256).
			WithUser(senderKeyPair).
			CreateFoundry()
		require.NoError(t, err)
		require.EqualValues(t, 1, sn)

		err = ch.DestroyTokensOnL2(sn, big.NewInt(1), senderKeyPair)
		testmisc.RequireErrorToBe(t, err, accounts.ErrNotEnoughFunds)
		ch.AssertL2AccountNativeToken(senderAgentID, &tokenID, big.NewInt(0))
		ch.AssertL2TotalNativeTokens(&tokenID, big.NewInt(0))
	})
	t.Run("max supply 100, mint_20, destroy_10", func(t *testing.T) {
		initTest()
		sn, tokenID, err := ch.NewFoundryParams(100).
			WithUser(senderKeyPair).
			CreateFoundry()
		require.NoError(t, err)
		require.EqualValues(t, 1, sn)

		out, err := ch.GetFoundryOutput(1)
		require.NoError(t, err)
		require.EqualValues(t, out.MustNativeTokenID(), tokenID)
		ch.AssertL2AccountNativeToken(senderAgentID, &tokenID, big.NewInt(0))
		ch.AssertL2TotalNativeTokens(&tokenID, big.NewInt(0))

		err = ch.SendFromL1ToL2AccountIotas(1000, ch.CommonAccount(), senderKeyPair)
		require.NoError(t, err)
		err = ch.MintTokens(sn, 20, senderKeyPair)
		require.NoError(t, err)
		ch.AssertL2AccountNativeToken(senderAgentID, &tokenID, 20)
		ch.AssertL2TotalNativeTokens(&tokenID, 20)

		err = ch.DestroyTokensOnL2(sn, 10, senderKeyPair)
		require.NoError(t, err)
		ch.AssertL2TotalNativeTokens(&tokenID, 10)
		ch.AssertL2AccountNativeToken(senderAgentID, &tokenID, 10)
	})
	t.Run("max supply 1000000, mint_1000000, destroy_1000000", func(t *testing.T) {
		initTest()
		sn, tokenID, err := ch.NewFoundryParams(1_000_000).
			WithUser(senderKeyPair).
			CreateFoundry()
		require.NoError(t, err)
		require.EqualValues(t, 1, sn)

		out, err := ch.GetFoundryOutput(1)
		require.NoError(t, err)
		require.EqualValues(t, out.MustNativeTokenID(), tokenID)
		ch.AssertL2AccountNativeToken(senderAgentID, &tokenID, 0)
		ch.AssertL2TotalNativeTokens(&tokenID, 0)

		err = ch.SendFromL1ToL2AccountIotas(1000, ch.CommonAccount(), senderKeyPair)
		require.NoError(t, err)
		err = ch.MintTokens(sn, 1_000_000, senderKeyPair)
		require.NoError(t, err)
		ch.AssertL2AccountNativeToken(senderAgentID, &tokenID, big.NewInt(1_000_000))
		ch.AssertL2TotalNativeTokens(&tokenID, big.NewInt(1_000_000))
		out, err = ch.GetFoundryOutput(1)
		require.NoError(t, err)
		require.True(t, big.NewInt(1_000_000).Cmp(out.CirculatingSupply) == 0)

		// FIXME bug iotago can't destroy foundry
		// err = destroyTokens(sn, big.NewInt(1000000))
		// require.NoError(t, err)
		// ch.AssertL2TotalNativeTokens(&tokenID, big.NewInt(0))
		// ch.AssertL2AccountNativeToken(userAgentID, &tokenID, big.NewInt(0))
		// out, err = ch.GetFoundryOutput(1)
		// require.NoError(t, err)
		// require.True(t, big.NewInt(0).Cmp(out.CirculatingSupply) == 0)
	})
	t.Run("10 foundries", func(t *testing.T) {
		initTest()

		for sn := uint32(1); sn <= 10; sn++ {
			var tag iotago.TokenTag
			copy(tag[:], util.Uint32To4Bytes(sn))
			snBack, tokenID, err := ch.NewFoundryParams(uint64(sn + 1)).
				WithUser(senderKeyPair).
				WithTag(&tag).
				CreateFoundry()
			require.NoError(t, err)
			require.EqualValues(t, int(sn), int(snBack))
			ch.AssertL2AccountNativeToken(senderAgentID, &tokenID, big.NewInt(0))
			ch.AssertL2TotalNativeTokens(&tokenID, big.NewInt(0))
		}
		// mint max supply from each
		for sn := uint32(1); sn <= 10; sn++ {
			err := ch.MintTokens(sn, sn+1, senderKeyPair)
			require.NoError(t, err)

			out, err := ch.GetFoundryOutput(sn)
			require.NoError(t, err)

			require.EqualValues(t, sn, out.SerialNumber)
			require.True(t, out.MaximumSupply.Cmp(big.NewInt(int64(sn+1))) == 0)
			require.True(t, out.CirculatingSupply.Cmp(big.NewInt(int64(sn+1))) == 0)
			tokenID := out.MustNativeTokenID()

			ch.AssertL2AccountNativeToken(senderAgentID, &tokenID, big.NewInt(int64(sn+1)))
			ch.AssertL2TotalNativeTokens(&tokenID, big.NewInt(int64(sn+1)))
		}
		// destroy 1 token of each tokenID
		for sn := uint32(1); sn <= 10; sn++ {
			err := ch.DestroyTokensOnL2(sn, big.NewInt(1), senderKeyPair)
			require.NoError(t, err)
		}
		// check balances
		for sn := uint32(1); sn <= 10; sn++ {
			out, err := ch.GetFoundryOutput(sn)
			require.NoError(t, err)

			require.EqualValues(t, sn, out.SerialNumber)
			require.True(t, out.MaximumSupply.Cmp(big.NewInt(int64(sn+1))) == 0)
			require.True(t, out.CirculatingSupply.Cmp(big.NewInt(int64(sn))) == 0)
			tokenID := out.MustNativeTokenID()

			ch.AssertL2AccountNativeToken(senderAgentID, &tokenID, big.NewInt(int64(sn)))
			ch.AssertL2TotalNativeTokens(&tokenID, big.NewInt(int64(sn)))
		}
	})
}

// TestFoundryValidation reveals bug in iota.go. Validation fails when whole supply is destroyed
func TestFoundryValidation(t *testing.T) {
	tokenID := tpkg.RandNativeToken().ID
	inSums := iotago.NativeTokenSum{
		tokenID: big.NewInt(1000000),
	}
	circSupplyChange := big.NewInt(-1000000)

	outSumsBad := iotago.NativeTokenSum{}
	outSumsGood := iotago.NativeTokenSum{tokenID: big.NewInt(0)}

	t.Run("fail", func(t *testing.T) {
		err := iotago.NativeTokenSumBalancedWithDiff(tokenID, inSums, outSumsBad, circSupplyChange)
		require.NoError(t, err)
	})
	t.Run("pass", func(t *testing.T) {
		err := iotago.NativeTokenSumBalancedWithDiff(tokenID, inSums, outSumsGood, circSupplyChange)
		require.NoError(t, err)
	})
}

func TestAccountBalances(t *testing.T) {
	env := solo.New(t)

	chainOwner, chainOwnerAddr := env.NewKeyPairWithFunds(env.NewSeedFromIndex(10))
	chainOwnerAgentID := iscp.NewAgentID(chainOwnerAddr, 0)

	sender, senderAddr := env.NewKeyPairWithFunds(env.NewSeedFromIndex(11))
	senderAgentID := iscp.NewAgentID(senderAddr, 0)

	l1Iotas := func(addr iotago.Address) uint64 { return env.L1AddressBalances(addr).Iotas }
	totalIotas := l1Iotas(chainOwnerAddr) + l1Iotas(senderAddr)

	chain := env.NewChain(chainOwner, "chain1")

	l2Iotas := func(agentID *iscp.AgentID) uint64 { return chain.L2AccountIotas(agentID) }
	totalGasFeeCharged := uint64(0)

	checkBalance := func(numReqs int) {
		require.EqualValues(t,
			totalIotas,
			l1Iotas(chainOwnerAddr)+l1Iotas(senderAddr)+l1Iotas(chain.ChainID.AsAddress()),
		)

		anchor, _ := chain.GetAnchorOutput()
		require.EqualValues(t, l1Iotas(chain.ChainID.AsAddress()), anchor.Deposit())

		require.LessOrEqual(t, len(chain.L2Accounts()), 3)

		bi := chain.GetLatestBlockInfo()

		require.EqualValues(t,
			anchor.Deposit(),
			bi.TotalIotasInL2Accounts+bi.TotalDustDeposit,
		)

		require.EqualValues(t,
			bi.TotalIotasInL2Accounts,
			l2Iotas(chainOwnerAgentID)+l2Iotas(senderAgentID)+l2Iotas(chain.CommonAccount()),
		)

		require.Equal(t, numReqs == 0, bi.GasFeeCharged == 0)
		totalGasFeeCharged += bi.GasFeeCharged
		require.EqualValues(t,
			l2Iotas(chain.CommonAccount()),
			totalGasFeeCharged,
		)

		require.EqualValues(t,
			solo.Saldo+totalGasFeeCharged-bi.TotalDustDeposit,
			l1Iotas(chainOwnerAddr)+l2Iotas(chainOwnerAgentID)+l2Iotas(chain.CommonAccount()),
		)
		require.EqualValues(t,
			solo.Saldo-totalGasFeeCharged,
			l1Iotas(senderAddr)+l2Iotas(senderAgentID),
		)
	}

	checkBalance(0)

	for i := 0; i < 5; i++ {
		_, err := chain.UploadBlob(sender, "field", fmt.Sprintf("dummy blob data #%d", i))
		require.NoError(t, err)

		checkBalance(i + 1)
	}
}

type testParams struct {
	env               *solo.Solo
	chainOwner        *cryptolib.KeyPair
	chainOwnerAddr    iotago.Address
	chainOwnerAgentID *iscp.AgentID
	user              *cryptolib.KeyPair
	userAddr          iotago.Address
	userAgentID       *iscp.AgentID
	ch                *solo.Chain
	req               *solo.CallParams
	sn                uint32
	tokenID           *iotago.NativeTokenID
}

func initTest(t *testing.T, initLoad ...uint64) *testParams {
	ret := &testParams{}
	ret.env = solo.New(t)

	ret.chainOwner, ret.chainOwnerAddr = ret.env.NewKeyPairWithFunds(ret.env.NewSeedFromIndex(10))
	ret.chainOwnerAgentID = iscp.NewAgentID(ret.chainOwnerAddr, 0)
	ret.user, ret.userAddr = ret.env.NewKeyPairWithFunds(ret.env.NewSeedFromIndex(11))
	ret.userAgentID = iscp.NewAgentID(ret.userAddr, 0)

	if len(initLoad) == 0 {
		ret.ch = ret.env.NewChain(ret.chainOwner, "chain1")
	} else {
		ret.ch, _, _ = ret.env.NewChainExt(ret.chainOwner, initLoad[0], "chain1")
	}
	ret.req = solo.NewCallParams(accounts.Contract.Name, accounts.FuncDeposit.Name)
	return ret
}

func (v *testParams) createFoundryAndMint(sch iotago.TokenScheme, tag *iotago.TokenTag, maxSupply, amount interface{}) (uint32, *iotago.NativeTokenID) {
	sn, tokenID, err := v.ch.NewFoundryParams(maxSupply).
		WithTokenScheme(sch).
		WithTag(tag).
		WithUser(v.user).
		CreateFoundry()
	require.NoError(v.env.T, err)
	// mint some tokens for the user
	err = v.ch.SendFromL1ToL2AccountIotas(1000, v.ch.CommonAccount(), v.user)
	require.NoError(v.env.T, err)
	err = v.ch.MintTokens(sn, amount, v.user)
	require.NoError(v.env.T, err)
	// check the balance of the user
	v.ch.AssertL2AccountNativeToken(v.userAgentID, &tokenID, amount)
	require.True(v.env.T, v.ch.L2AccountIotas(v.userAgentID) > 100) // must be some coming from dust deposits
	return sn, &tokenID
}

func TestDepositIotas(t *testing.T) {
	// the test check how request transaction construction functions adjust iotas to the minimum needed for the
	// dust deposit. If byte cost is 185, anything below that fill be topped up to 185, above that no adjustment is needed
	for _, addIotas := range []uint64{0, 50, 150, 200, 1000} {
		t.Run("add iotas "+strconv.Itoa(int(addIotas)), func(t *testing.T) {
			v := initTest(t)
			v.req = v.req.AddAssetsIotas(addIotas)
			tx, _, err := v.ch.PostRequestSyncTx(v.req, v.user)
			require.NoError(t, err)

			byteCost := tx.Essence.Outputs[0].VByteCost(v.env.RentStructure(), nil)
			t.Logf("byteCost = %d", byteCost)

			// here we calculate what is expected
			expected := addIotas
			if expected < byteCost {
				expected = byteCost
			}
			v.ch.AssertL2AccountIotas(v.userAgentID, expected)
		})
	}
}

// initWithdrawTest creates foundry with 1_000_000 of max supply and mint 100 tokens to user's account
func initWithdrawTest(t *testing.T, initLoad ...uint64) *testParams {
	v := initTest(t, initLoad...)
	// create foundry and mint 100 tokens
	v.sn, v.tokenID = v.createFoundryAndMint(nil, nil, 1_000_000, 100)
	// prepare request parameters to withdraw everything what is in the account
	// do not run the request yet
	v.req = solo.NewCallParams("accounts", "withdraw").
		AddAssetsIotas(1000).
		WithGasBudget(2000)
	v.printBalances("BEGIN")
	return v
}

func (v *testParams) printBalances(prefix string) {
	v.env.T.Logf("%s: user L1 iotas: %d", prefix, v.env.L1IotaBalance(v.userAddr))
	v.env.T.Logf("%s: user L1 tokens: %s : %d", prefix, v.tokenID, v.env.L1NativeTokenBalance(v.userAddr, v.tokenID))
	v.env.T.Logf("%s: user L2: %s", prefix, v.ch.L2AccountAssets(v.userAgentID))
	v.env.T.Logf("%s: common account L2: %s", prefix, v.ch.L2CommonAccountAssets())
}

func TestWithdrawDepositNativeTokens(t *testing.T) {
	t.Run("withdraw with empty", func(t *testing.T) {
		v := initWithdrawTest(t)
		_, err := v.ch.PostRequestSync(v.req, v.user)
		testmisc.RequireErrorToBe(t, err, "can't be empty")
	})
	t.Run("withdraw not enough for dust", func(t *testing.T) {
		v := initWithdrawTest(t)
		v.req.AddNativeTokensAllowanceVect(&iotago.NativeToken{
			ID:     *v.tokenID,
			Amount: big.NewInt(100),
		})
		_, err := v.ch.PostRequestSync(v.req, v.user)
		testmisc.RequireErrorToBe(t, err, accounts.ErrNotEnoughIotasForDustDeposit)
	})
	t.Run("withdraw almost all", func(t *testing.T) {
		v := initWithdrawTest(t)
		// we want to withdraw as many iotas as possible, so we add 300 because some more will come
		// with assets attached to the 'withdraw' request. However, withdraw all is not possible due to gas
		toWithdraw := v.ch.L2AccountAssets(v.userAgentID).AddIotas(200)
		t.Logf("assets to withdraw: %s", toWithdraw.String())
		// withdraw all tokens to L1, but we do not add iotas to allowance, so not enough for dust
		v.req.AddAllowance(toWithdraw)
		v.req.AddAssetsIotas(1000)
		_, err := v.ch.PostRequestSync(v.req, v.user)
		require.NoError(t, err)
		v.printBalances("END")
	})
	t.Run("mint withdraw destroy fail", func(t *testing.T) {
		v := initWithdrawTest(t)
		allSenderAssets := v.ch.L2AccountAssets(v.userAgentID)
		v.req.AddAllowance(allSenderAssets)
		v.req.AddAssetsIotas(1000)
		_, err := v.ch.PostRequestSync(v.req, v.user)
		require.NoError(t, err)

		v.printBalances("AFTER MINT")
		v.env.AssertL1NativeTokens(v.userAddr, v.tokenID, 100)

		// should fail because those tokens are not on the user's on chain account
		err = v.ch.DestroyTokensOnL2(v.sn, big.NewInt(50), v.user)
		testmisc.RequireErrorToBe(t, err, accounts.ErrNotEnoughFunds)
		v.env.AssertL1NativeTokens(v.userAddr, v.tokenID, big.NewInt(100))
		v.printBalances("AFTER DESTROY")
	})
	t.Run("mint withdraw destroy success 1", func(t *testing.T) {
		v := initWithdrawTest(t)

		allSenderAssets := v.ch.L2AccountAssets(v.userAgentID)
		v.req.AddAllowance(allSenderAssets)
		v.req.AddAssetsIotas(1000)
		_, err := v.ch.PostRequestSync(v.req, v.user)
		require.NoError(t, err)
		v.printBalances("AFTER MINT")
		v.env.AssertL1NativeTokens(v.userAddr, v.tokenID, 100)
		v.ch.AssertL2AccountNativeToken(v.userAgentID, v.tokenID, 0)

		err = v.ch.DepositAssets(iscp.NewEmptyAssets().AddNativeTokens(*v.tokenID, 50), v.user)
		require.NoError(t, err)
		v.env.AssertL1NativeTokens(v.userAddr, v.tokenID, 50)
		v.ch.AssertL2AccountNativeToken(v.userAgentID, v.tokenID, 50)
		v.ch.AssertL2TotalNativeTokens(v.tokenID, 50)
		v.printBalances("AFTER DEPOSIT")

		err = v.ch.DestroyTokensOnL2(v.sn, 49, v.user)
		require.NoError(t, err)
		v.ch.AssertL2AccountNativeToken(v.userAgentID, v.tokenID, 1)
		v.env.AssertL1NativeTokens(v.userAddr, v.tokenID, 50)
		v.printBalances("AFTER DESTROY")
	})
	t.Run("unwrap use case", func(t *testing.T) {
		v := initWithdrawTest(t)
		allSenderAssets := v.ch.L2AccountAssets(v.userAgentID)
		v.req.AddAllowance(allSenderAssets)
		v.req.AddAssetsIotas(1000)
		_, err := v.ch.PostRequestSync(v.req, v.user)
		require.NoError(t, err)
		v.printBalances("AFTER MINT")
		v.env.AssertL1NativeTokens(v.userAddr, v.tokenID, 100)
		v.ch.AssertL2AccountNativeToken(v.userAgentID, v.tokenID, 0)

		err = v.ch.DepositAssets(iscp.NewEmptyAssets().AddNativeTokens(*v.tokenID, 1), v.user)
		require.NoError(t, err)
		v.printBalances("AFTER DEPOSIT 1")

		// without deposit
		err = v.ch.DestroyTokensOnL1(v.tokenID, 49, v.user)
		require.NoError(t, err)
		v.printBalances("AFTER DESTROY")
		v.ch.AssertL2AccountNativeToken(v.userAgentID, v.tokenID, 1)
		v.env.AssertL1NativeTokens(v.userAddr, v.tokenID, 50)
	})
	t.Run("unwrap use case", func(t *testing.T) {
		v := initWithdrawTest(t)
		allSenderAssets := v.ch.L2AccountAssets(v.userAgentID)
		v.req.AddAllowance(allSenderAssets)
		v.req.AddAssetsIotas(1000)
		_, err := v.ch.PostRequestSync(v.req, v.user)
		require.NoError(t, err)
		v.printBalances("AFTER MINT")
		v.env.AssertL1NativeTokens(v.userAddr, v.tokenID, 100)
		v.ch.AssertL2AccountNativeToken(v.userAgentID, v.tokenID, 0)

		// without deposit
		err = v.ch.DestroyTokensOnL1(v.tokenID, 49, v.user)
		require.NoError(t, err)
		v.printBalances("AFTER DESTROY")
		v.ch.AssertL2AccountNativeToken(v.userAgentID, v.tokenID, 0)
		v.env.AssertL1NativeTokens(v.userAddr, v.tokenID, 51)
	})
	t.Run("mint withdraw destroy fail", func(t *testing.T) {
		v := initWithdrawTest(t)
		allSenderAssets := v.ch.L2AccountAssets(v.userAgentID)
		v.req.AddAllowance(allSenderAssets)
		v.req.AddAssetsIotas(1000)
		_, err := v.ch.PostRequestSync(v.req, v.user)
		require.NoError(t, err)

		v.printBalances("AFTER MINT")
		v.env.AssertL1NativeTokens(v.userAddr, v.tokenID, 100)
		v.ch.AssertL2AccountNativeToken(v.userAgentID, v.tokenID, 0)

		err = v.ch.DepositAssets(iscp.NewEmptyAssets().AddNativeTokens(*v.tokenID, 50), v.user)
		require.NoError(t, err)
		v.env.AssertL1NativeTokens(v.userAddr, v.tokenID, 50)
		v.ch.AssertL2AccountNativeToken(v.userAgentID, v.tokenID, 50)
		v.ch.AssertL2TotalNativeTokens(v.tokenID, 50)
		v.printBalances("AFTER DEPOSIT")

		err = v.ch.DestroyTokensOnL2(v.sn, 50, v.user)
		require.NoError(t, err)
		v.ch.AssertL2AccountNativeToken(v.userAgentID, v.tokenID, 0)
		v.env.AssertL1NativeTokens(v.userAddr, v.tokenID, 50)
	})
}

func TestTransferAndHarvest(t *testing.T) {
	// initializes it all and prepares withdraw request, does not post it
	v := initWithdrawTest(t, 10_1000)
	dustCosts := transaction.NewDepositEstimate(v.env.RentStructure())
	commonAssets := v.ch.L2CommonAccountAssets()
	require.True(t, commonAssets.Iotas+dustCosts.AnchorOutput > 10_000)
	require.EqualValues(t, 0, len(commonAssets.Tokens))

	v.ch.AssertL2AccountNativeToken(v.userAgentID, v.tokenID, 100)

	// move minted tokens from user to the common account on-chain
	err := v.ch.SendFromL2ToL2AccountNativeTokens(*v.tokenID, v.ch.CommonAccount(), 50, v.user)
	require.NoError(t, err)
	// now we have 50 tokens on common account
	v.ch.AssertL2AccountNativeToken(v.ch.CommonAccount(), v.tokenID, 50)
	// no native tokens for chainOwner on L1
	v.env.AssertL1NativeTokens(v.chainOwnerAddr, v.tokenID, 0)

	v.req = solo.NewCallParams("accounts", "harvest").
		WithGasBudget(1000)
	receipt, _, err := v.ch.PostRequestSyncReceipt(v.req, v.chainOwner)
	require.NoError(t, err)

	t.Logf("receipt from the 'harvest' tx: %s", receipt)

	// now we have 0 tokens on common account
	v.ch.AssertL2AccountNativeToken(v.ch.CommonAccount(), v.tokenID, 0)
	// 50 native tokens for chain on L2
	v.ch.AssertL2AccountNativeToken(v.chainOwnerAgentID, v.tokenID, 50)

	commonAssets = v.ch.L2CommonAccountAssets()
	// in the common account should have left minimum plus gas fee from the last request
	require.EqualValues(t, accounts.MinimumIotasOnCommonAccount+receipt.GasFeeCharged, commonAssets.Iotas)
	require.EqualValues(t, 0, len(commonAssets.Tokens))
}

func TestFoundryDestroy(t *testing.T) {
	t.Run("destroy existing", func(t *testing.T) {
		v := initTest(t)
		sn, _, err := v.ch.NewFoundryParams(1_000_000).
			WithUser(v.user).
			CreateFoundry()
		require.NoError(t, err)

		err = v.ch.DestroyFoundry(sn, v.user)
		require.NoError(t, err)
		_, err = v.ch.GetFoundryOutput(sn)
		testmisc.RequireErrorToBe(t, err, "does not exist")
	})
	t.Run("destroy fail", func(t *testing.T) {
		v := initTest(t)
		err := v.ch.DestroyFoundry(2, v.user)
		testmisc.RequireErrorToBe(t, err, "not controlled by the caller")
	})
}

func TestTransferPartialAssets(t *testing.T) {
	e := initTest(t)
	// setup a chain with some iotas and native tokens for user1
	sn, tokenID, err := e.ch.NewFoundryParams(10).
		WithUser(e.user).
		CreateFoundry()
	require.NoError(t, err)
	require.EqualValues(t, 1, int(sn))

	// deposit iotas for the chain owner (needed for L1 dust byte cost to mint tokens)
	err = e.ch.SendFromL1ToL2AccountIotas(10000, e.ch.CommonAccount(), e.chainOwner)
	require.NoError(t, err)
	err = e.ch.SendFromL1ToL2AccountIotas(10000, e.userAgentID, e.user)
	require.NoError(t, err)

	err = e.ch.MintTokens(sn, big.NewInt(10), e.user)
	require.NoError(t, err)

	e.ch.AssertL2AccountNativeToken(e.userAgentID, &tokenID, big.NewInt(10))
	e.ch.AssertL2TotalNativeTokens(&tokenID, big.NewInt(10))

	// send funds to user2
	par := dict.New()
	user2, user2Addr := e.env.NewKeyPairWithFunds(e.env.NewSeedFromIndex(100))
	user2AgentID := iscp.NewAgentID(user2Addr, 0)

	// deposit 1 iota to "create account" for user2 // TODO maybe remove if account creation is not needed
	e.ch.AssertL2AccountIotas(user2AgentID, 0)
	e.ch.SendFromL1ToL2AccountIotas(300, user2AgentID, user2)
	e.ch.AssertL2AccountIotas(user2AgentID, 300)
	// -----------------------------

	par.Set(accounts.ParamAgentID, codec.EncodeAgentID(user2AgentID))

	err = e.ch.SendFromL2ToL2Account(
		iscp.NewAssets(300, iotago.NativeTokens{
			&iotago.NativeToken{
				ID:     tokenID,
				Amount: big.NewInt(9),
			},
		}),
		user2AgentID,
		e.user,
	)
	require.NoError(t, err)

	// assert that balances are correct
	e.ch.AssertL2AccountNativeToken(e.userAgentID, &tokenID, big.NewInt(1))
	e.ch.AssertL2AccountNativeToken(user2AgentID, &tokenID, big.NewInt(9))
	e.ch.AssertL2AccountIotas(user2AgentID, 600) // TODO should be 300 if not needed to create account manually
	e.ch.AssertL2TotalNativeTokens(&tokenID, big.NewInt(10))
}

// TestCirculatingSupplyBurn belongs to iota.go
func TestCirculatingSupplyBurn(t *testing.T) {
	const OneMi = 1_000_000

	_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
	aliasIdent1 := tpkg.RandAliasAddress()

	tokenTag := tpkg.Rand12ByteArray()

	inputIDs := tpkg.RandOutputIDs(3)
	inputs := iotago.OutputSet{
		inputIDs[0]: &iotago.ExtendedOutput{
			Address: ident1,
			Amount:  OneMi,
		},
		inputIDs[1]: &iotago.AliasOutput{
			Amount:               OneMi,
			NativeTokens:         nil,
			AliasID:              aliasIdent1.AliasID(),
			StateController:      ident1,
			GovernanceController: ident1,
			StateIndex:           1,
			StateMetadata:        nil,
			FoundryCounter:       1,
			Blocks:               nil,
		},
		inputIDs[2]: &iotago.FoundryOutput{
			Address:           aliasIdent1,
			Amount:            OneMi,
			NativeTokens:      nil,
			SerialNumber:      1,
			TokenTag:          tokenTag,
			CirculatingSupply: big.NewInt(50),
			MaximumSupply:     big.NewInt(50),
			TokenScheme:       &iotago.SimpleTokenScheme{},
			Blocks:            nil,
		},
	}

	// set input ExtendedOutput NativeToken to 50 which get burned
	foundryNativeTokenID := inputs[inputIDs[2]].(*iotago.FoundryOutput).MustNativeTokenID()
	inputs[inputIDs[0]].(*iotago.ExtendedOutput).NativeTokens = iotago.NativeTokens{
		{
			ID:     foundryNativeTokenID,
			Amount: new(big.Int).SetInt64(50),
		},
	}

	essence := &iotago.TransactionEssence{
		Inputs: inputIDs.UTXOInputs(),
		Outputs: iotago.Outputs{
			&iotago.AliasOutput{
				Amount:               OneMi,
				NativeTokens:         nil,
				AliasID:              aliasIdent1.AliasID(),
				StateController:      ident1,
				GovernanceController: ident1,
				StateIndex:           1,
				StateMetadata:        nil,
				FoundryCounter:       1,
				Blocks:               nil,
			},
			&iotago.FoundryOutput{
				Address:      aliasIdent1,
				Amount:       2 * OneMi,
				NativeTokens: nil,
				SerialNumber: 1,
				TokenTag:     tokenTag,
				// burn supply by -50
				CirculatingSupply: big.NewInt(0),
				MaximumSupply:     big.NewInt(50),
				TokenScheme:       &iotago.SimpleTokenScheme{},
				Blocks:            nil,
			},
		},
	}

	sigs, err := essence.Sign(ident1AddrKeys)
	require.NoError(t, err)

	tx := &iotago.Transaction{
		Essence: essence,
		UnlockBlocks: iotago.UnlockBlocks{
			&iotago.SignatureUnlockBlock{Signature: sigs[0]},
			&iotago.ReferenceUnlockBlock{Reference: 0},
			&iotago.AliasUnlockBlock{Reference: 1},
		},
	}

	require.NoError(t, tx.SemanticallyValidate(&iotago.SemanticValidationContext{
		ExtParas:   nil,
		WorkingSet: nil,
	}, inputs))
}
