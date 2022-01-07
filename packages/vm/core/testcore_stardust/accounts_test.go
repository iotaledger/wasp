package testcore

import (
	"fmt"
	"math/big"
	"strconv"
	"testing"

	"github.com/iotaledger/wasp/packages/testutil/testmisc"

	"github.com/ethereum/go-ethereum/accounts/abi"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/solo"
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

		err = ch.MintTokens(sn, big.NewInt(1), senderKeyPair)
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

		err = ch.MintTokens(sn, big.NewInt(2), senderKeyPair)
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

		err = ch.MintTokens(sn, big.NewInt(500), senderKeyPair)
		require.NoError(t, err)
		ch.AssertL2AccountNativeToken(senderAgentID, &tokenID, big.NewInt(500))
		ch.AssertL2TotalNativeTokens(&tokenID, big.NewInt(500))

		err = ch.MintTokens(sn, big.NewInt(500), senderKeyPair)
		require.NoError(t, err)
		ch.AssertL2AccountNativeToken(senderAgentID, &tokenID, big.NewInt(1000))
		ch.AssertL2TotalNativeTokens(&tokenID, big.NewInt(1000))

		err = ch.MintTokens(sn, big.NewInt(1), senderKeyPair)
		testmisc.RequireErrorToBe(t, err, vmtxbuilder.ErrNativeTokenSupplyOutOffBounds)

		ch.AssertL2AccountNativeToken(senderAgentID, &tokenID, big.NewInt(1000))
		ch.AssertL2TotalNativeTokens(&tokenID, big.NewInt(1000))
	})
	t.Run("max supply MaxUint256, mintTokens MaxUint256_1", func(t *testing.T) {
		initTest()
		sn, tokenID, err := ch.NewFoundryParams(abi.MaxUint256).
			WithUser(senderKeyPair).
			CreateFoundry()
		require.NoError(t, err)
		require.EqualValues(t, 1, sn)

		err = ch.MintTokens(sn, abi.MaxUint256, senderKeyPair)
		require.NoError(t, err)
		ch.AssertL2AccountNativeToken(senderAgentID, &tokenID, abi.MaxUint256)

		err = ch.MintTokens(sn, big.NewInt(1), senderKeyPair)
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

		err = ch.DestroyTokens(sn, big.NewInt(1), senderKeyPair)
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

		err = ch.MintTokens(sn, big.NewInt(20), senderKeyPair)
		require.NoError(t, err)
		ch.AssertL2AccountNativeToken(senderAgentID, &tokenID, big.NewInt(20))
		ch.AssertL2TotalNativeTokens(&tokenID, big.NewInt(20))

		err = ch.DestroyTokens(sn, big.NewInt(10), senderKeyPair)
		require.NoError(t, err)
		ch.AssertL2TotalNativeTokens(&tokenID, big.NewInt(10))
		ch.AssertL2AccountNativeToken(senderAgentID, &tokenID, big.NewInt(10))
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
		ch.AssertL2AccountNativeToken(senderAgentID, &tokenID, big.NewInt(0))
		ch.AssertL2TotalNativeTokens(&tokenID, big.NewInt(0))

		err = ch.MintTokens(sn, big.NewInt(1_000_000), senderKeyPair)
		require.NoError(t, err)
		ch.AssertL2AccountNativeToken(senderAgentID, &tokenID, big.NewInt(1_000_000))
		ch.AssertL2TotalNativeTokens(&tokenID, big.NewInt(1_000_000))
		out, err = ch.GetFoundryOutput(1)
		require.NoError(t, err)
		require.True(t, big.NewInt(1_000_000).Cmp(out.CirculatingSupply) == 0)

		//err = destroyTokens(sn, big.NewInt(1000000)) // <<<<<<<< fails TODO
		//require.NoError(t, err)
		//ch.AssertL2TotalNativeTokens(&tokenID, big.NewInt(0))
		//ch.AssertL2AccountNativeToken(userAgentID, &tokenID, big.NewInt(0))
		//out, err = ch.GetFoundryOutput(1)
		//require.NoError(t, err)
		//require.True(t, big.NewInt(0).Cmp(out.CirculatingSupply) == 0)
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
			err := ch.MintTokens(sn, big.NewInt(int64(sn+1)), senderKeyPair)
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
			err := ch.DestroyTokens(sn, big.NewInt(1), senderKeyPair)
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
		require.Error(t, err) // <<<<<<<<<<<<<<<<<<< TODO wrong
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
	env         *solo.Solo
	chainOwner  *cryptolib.KeyPair
	user        *cryptolib.KeyPair
	userAddr    iotago.Address
	userAgentID *iscp.AgentID
	ch          *solo.Chain
	req         *solo.CallParams
}

func initTest(t *testing.T) *testParams {
	ret := &testParams{}
	ret.env = solo.New(t)

	ret.chainOwner, _ = ret.env.NewKeyPairWithFunds(ret.env.NewSeedFromIndex(10))
	//chainOwnerAgentID := iscp.NewAgentID(chainOwnerAddr, 0)

	ret.user, ret.userAddr = ret.env.NewKeyPairWithFunds(ret.env.NewSeedFromIndex(11))
	ret.userAgentID = iscp.NewAgentID(ret.userAddr, 0)

	ret.ch = ret.env.NewChain(ret.chainOwner, "chain1")
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
	err = v.ch.MintTokens(sn, amount, v.user)
	require.NoError(v.env.T, err)
	// check the balance of the user
	v.ch.AssertL2AccountNativeToken(v.userAgentID, &tokenID, amount)
	require.True(v.env.T, v.ch.L2AccountIotas(v.userAgentID) > 100) // must be some coming from dust deposits
	v.env.T.Logf("user has on-chain: %d", v.ch.L2AccountIotas(v.userAgentID))
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

func TestWithdrawDepositNativeTokens(t *testing.T) {
	t.Run("withdraw with empty", func(t *testing.T) {
		v := initTest(t)
		v.createFoundryAndMint(nil, nil, big.NewInt(1_000_000), big.NewInt(100))
		// withdraw all tokens to L1
		req := solo.NewCallParams("accounts", "withdraw").
			AddAssetsIotas(1000).
			WithGasBudget(2000)
		_, err := v.ch.PostRequestSync(req, v.user)
		testmisc.RequireErrorToBe(t, err, "can't be empty")
	})
	t.Run("withdraw not enough for dust", func(t *testing.T) {
		v := initTest(t)
		_, tokenID := v.createFoundryAndMint(nil, nil, big.NewInt(1_000_000), big.NewInt(100))
		// withdraw all tokens to L1, but we do not add iotas to allowance, so not enough for dust
		req := solo.NewCallParams("accounts", "withdraw").
			AddAssetsIotas(2000).
			WithGasBudget(2000).
			AddNativeTokensAllowance(&iotago.NativeToken{
				ID:     *tokenID,
				Amount: big.NewInt(100),
			})
		_, err := v.ch.PostRequestSync(req, v.user)
		testmisc.RequireErrorToBe(t, err, accounts.ErrNotEnoughIotasForDustDeposit)
	})
	t.Run("withdraw almost all", func(t *testing.T) {
		v := initTest(t)
		v.createFoundryAndMint(nil, nil, big.NewInt(1_000_000), big.NewInt(100))
		allSenderAssets := v.ch.L2AccountBalances(v.userAgentID)
		t.Logf("user L1 iotas: %d", v.env.L1IotaBalance(v.userAddr))

		// we want withdraw as much iotas as possible, however, all is not possible due to gas
		allSenderAssets.AddIotas(300)
		// withdraw all tokens to L1, but we do not add iotas to allowance, so not enough for dust
		req := solo.NewCallParams("accounts", "withdraw").
			AddAssetsIotas(2000).
			WithGasBudget(2000).
			AddAllowance(allSenderAssets)
		_, err := v.ch.PostRequestSync(req, v.user)
		require.NoError(t, err)
		t.Logf("user on L2 has: %s", v.ch.L2AccountBalances(v.userAgentID))
		t.Logf("user on L1 has: %s", v.env.L1AddressBalances(v.userAddr))
		t.Logf("L2 common account has: %s", v.ch.L2CommonAccountBalances())
	})
	t.Run("mint withdraw destroy fail", func(t *testing.T) {
		v := initTest(t)
		sn, tokenID := v.createFoundryAndMint(nil, nil, big.NewInt(1000), big.NewInt(100))
		allSenderAssets := v.ch.L2AccountBalances(v.userAgentID)
		t.Logf("user L1 iotas: %d", v.env.L1IotaBalance(v.userAddr))

		// withdraw all tokens to L1, but we do not add iotas to allowance, so not enough for dust
		req := solo.NewCallParams("accounts", "withdraw").
			AddAssetsIotas(2000).
			WithGasBudget(2000).
			AddAllowance(allSenderAssets)
		_, err := v.ch.PostRequestSync(req, v.user)
		require.NoError(t, err)
		t.Logf("user on L2 has: %s", v.ch.L2AccountBalances(v.userAgentID))
		t.Logf("user on L1 has: %s", v.env.L1AddressBalances(v.userAddr))
		t.Logf("L2 common account has: %s", v.ch.L2CommonAccountBalances())
		v.env.AssertL1NativeTokens(v.userAddr, tokenID, big.NewInt(100))

		// should fail because those tokens are not on the user's on chain account
		err = v.ch.DestroyTokens(sn, big.NewInt(50), v.user)
		testmisc.RequireErrorToBe(t, err, accounts.ErrNotEnoughFunds)
		v.env.AssertL1NativeTokens(v.userAddr, tokenID, big.NewInt(100))
	})
	t.Run("mint withdraw destroy success", func(t *testing.T) {
		v := initTest(t)
		sn, tokenID := v.createFoundryAndMint(nil, nil, big.NewInt(1000), big.NewInt(100))
		allSenderAssets := v.ch.L2AccountBalances(v.userAgentID)
		t.Logf("user L1 iotas: %d", v.env.L1IotaBalance(v.userAddr))

		v.env.AssertL1NativeTokens(v.userAddr, tokenID, big.NewInt(0))
		v.ch.AssertL2AccountNativeToken(v.userAgentID, tokenID, 100)

		// withdraw all tokens to L1, but we do not add iotas to allowance, so not enough for dust
		req := solo.NewCallParams("accounts", "withdraw").
			AddAssetsIotas(2000).
			WithGasBudget(2000).
			AddAllowance(allSenderAssets)
		_, err := v.ch.PostRequestSync(req, v.user)
		require.NoError(t, err)
		t.Logf("user on L2 has: %s", v.ch.L2AccountBalances(v.userAgentID))
		t.Logf("user on L1 has: %s", v.env.L1AddressBalances(v.userAddr))
		t.Logf("L2 common account has: %s", v.ch.L2CommonAccountBalances())
		v.env.AssertL1NativeTokens(v.userAddr, tokenID, big.NewInt(100))
		v.ch.AssertL2AccountNativeToken(v.userAgentID, tokenID, 0)

		err = v.ch.DepositAssets(iscp.NewEmptyAssets().AddNativeTokens(*tokenID, 50), v.user)
		v.env.AssertL1NativeTokens(v.userAddr, tokenID, 50)
		v.ch.AssertL2AccountNativeToken(v.userAgentID, tokenID, 50)
		v.ch.AssertL2TotalNativeTokens(tokenID, 50)

		err = v.ch.DestroyTokens(sn, 50, v.user)
		require.NoError(t, err)
		v.env.AssertL1NativeTokens(v.userAddr, tokenID, big.NewInt(100))
	})

	//t.Run("add many native tokens", func(t *testing.T) {
	//	for _, genTokens := range []int{1, 5, 10}  {
	//		t.Run("iotas+nt "+strconv.Itoa(genTokens), func(t *testing.T) {
	//			v := initTest(t)
	//			_, err := v.env.L1Ledger().GetFundsFromFaucet(v.userAddr)
	//			require.NoError(t, err)
	//
	//			expected := uint64(i * 50)
	//			v.req = v.req.AddAssetsIotas(expected)
	//			rndTokens := tpkg.RandSortNativeTokens(i)
	//			v.req = v.req.AddAssetsNativeTokens(rndTokens...)
	//			tx, _, err := v.ch.PostRequestSyncTx(v.req, v.user)
	//			require.NoError(t, err)
	//
	//			byteCost := tx.Essence.Outputs[0].VByteCost(v.env.RentStructure(), nil)
	//			t.Logf("byteCost = %d", byteCost)
	//			if expected < byteCost {
	//				expected = byteCost
	//			}
	//			v.ch.AssertL2AccountIotas(v.userAgentID, expected)
	//		})
	//	}
	//
	//})
}
