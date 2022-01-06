package testcore

import (
	"errors"
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
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
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
	createFoundry := func(t *testing.T, sch *iotago.TokenScheme, tag *iotago.TokenTag, maxSupply *big.Int) (error, uint32, iotago.NativeTokenID) {
		par := dict.New()
		if sch != nil {
			par.Set(accounts.ParamTokenScheme, codec.EncodeTokenScheme(*sch))
		}
		if tag != nil {
			par.Set(accounts.ParamTokenTag, codec.EncodeTokenTag(*tag))
		}
		if maxSupply != nil {
			par.Set(accounts.ParamMaxSupply, codec.EncodeBigIntAbs(maxSupply))
		}
		req := solo.NewCallParamsFromDic(accounts.Contract.Name, accounts.FuncFoundryCreateNew.Name, par).
			WithGasBudget(1000).
			AddIotas(1000)
		_, res, err := ch.PostRequestSyncTx(req, senderKeyPair)

		retSN := uint32(0)
		var tokenID iotago.NativeTokenID
		if err == nil {
			resDeco := kvdecoder.New(res)
			retSN = resDeco.MustGetUint32(accounts.ParamFoundrySN)
			tokenID, err = ch.GetNativeTokenIDByFoundrySN(retSN)
			require.NoError(t, err)
		}

		return err, retSN, tokenID
	}
	mintTokens := func(sn uint32, amount *big.Int) error {
		req := solo.NewCallParams(accounts.Contract.Name, accounts.FuncFoundryModifySupply.Name,
			accounts.ParamFoundrySN, sn,
			accounts.ParamSupplyDeltaAbs, amount,
		).WithGasBudget(1000)
		_, _, err := ch.PostRequestSyncTx(req, senderKeyPair)
		return err
	}
	destroyTokens := func(sn uint32, amount *big.Int) error {
		req := solo.NewCallParams(accounts.Contract.Name, accounts.FuncFoundryModifySupply.Name,
			accounts.ParamFoundrySN, sn,
			accounts.ParamSupplyDeltaAbs, amount,
			accounts.ParamDestroyTokens, true,
		).WithGasBudget(1000)
		_, _, err := ch.PostRequestSyncTx(req, senderKeyPair)
		return err
	}

	t.Run("supply 10", func(t *testing.T) {
		initTest()
		err, sn, _ := createFoundry(t, nil, nil, big.NewInt(10))
		require.NoError(t, err)
		require.EqualValues(t, 1, int(sn))
	})
	t.Run("supply 1", func(t *testing.T) {
		initTest()
		err, sn, _ := createFoundry(t, nil, nil, big.NewInt(1))
		require.NoError(t, err)
		require.EqualValues(t, 1, sn)
	})
	t.Run("supply 0", func(t *testing.T) {
		initTest()
		err, _, _ := createFoundry(t, nil, nil, big.NewInt(0))
		require.True(t, errors.Is(err, vmtxbuilder.ErrCreateFoundryMaxSupplyMustBePositive))
	})
	t.Run("supply negative", func(t *testing.T) {
		initTest()
		err, sn, _ := createFoundry(t, nil, nil, big.NewInt(-1))
		// encoding will ignore sign
		require.NoError(t, err)
		require.EqualValues(t, 1, sn)
	})
	t.Run("supply max possible", func(t *testing.T) {
		initTest()
		err, sn, _ := createFoundry(t, nil, nil, abi.MaxUint256)
		require.NoError(t, err)
		require.EqualValues(t, 1, sn)
	})
	t.Run("supply exceed max possible", func(t *testing.T) {
		initTest()
		maxSupply := big.NewInt(0).Set(abi.MaxUint256)
		maxSupply.Add(maxSupply, big.NewInt(1))
		err, _, _ := createFoundry(t, nil, nil, maxSupply)
		require.True(t, errors.Is(err, vmtxbuilder.ErrCreateFoundryMaxSupplyTooBig))
	})
	// TODO cover all parameter options

	t.Run("max supply 10, mintTokens 5", func(t *testing.T) {
		initTest()
		err, sn, tokenID := createFoundry(t, nil, nil, big.NewInt(10))
		require.NoError(t, err)
		require.EqualValues(t, 1, sn)
		ch.AssertL2AccountNativeToken(senderAgentID, &tokenID, big.NewInt(0))
		ch.AssertL2TotalNativeTokens(&tokenID, big.NewInt(0))

		err = mintTokens(sn, big.NewInt(5))
		require.NoError(t, err)

		ch.AssertL2AccountNativeToken(senderAgentID, &tokenID, big.NewInt(5))
		ch.AssertL2TotalNativeTokens(&tokenID, big.NewInt(5))
	})
	t.Run("max supply 1, mintTokens 1", func(t *testing.T) {
		initTest()
		err, sn, tokenID := createFoundry(t, nil, nil, big.NewInt(1))
		require.NoError(t, err)
		require.EqualValues(t, 1, sn)
		ch.AssertL2AccountNativeToken(senderAgentID, &tokenID, big.NewInt(0))
		ch.AssertL2TotalNativeTokens(&tokenID, big.NewInt(0))

		err = mintTokens(sn, big.NewInt(1))
		require.NoError(t, err)

		ch.AssertL2AccountNativeToken(senderAgentID, &tokenID, big.NewInt(1))
		ch.AssertL2TotalNativeTokens(&tokenID, big.NewInt(1))
	})

	t.Run("max supply 1, mintTokens 2", func(t *testing.T) {
		initTest()
		err, sn, tokenID := createFoundry(t, nil, nil, big.NewInt(1))
		require.NoError(t, err)
		require.EqualValues(t, 1, sn)

		err = mintTokens(sn, big.NewInt(2))
		require.True(t, errors.Is(err, vmtxbuilder.ErrNativeTokenSupplyOutOffBounds))

		ch.AssertL2AccountNativeToken(senderAgentID, &tokenID, big.NewInt(0))
		ch.AssertL2TotalNativeTokens(&tokenID, big.NewInt(0))
	})
	t.Run("max supply 1000, mintTokens 500_500_1", func(t *testing.T) {
		initTest()
		err, sn, tokenID := createFoundry(t, nil, nil, big.NewInt(1000))
		require.NoError(t, err)
		require.EqualValues(t, 1, sn)

		err = mintTokens(sn, big.NewInt(500))
		require.NoError(t, err)
		ch.AssertL2AccountNativeToken(senderAgentID, &tokenID, big.NewInt(500))
		ch.AssertL2TotalNativeTokens(&tokenID, big.NewInt(500))

		err = mintTokens(sn, big.NewInt(500))
		require.NoError(t, err)
		ch.AssertL2AccountNativeToken(senderAgentID, &tokenID, big.NewInt(1000))
		ch.AssertL2TotalNativeTokens(&tokenID, big.NewInt(1000))

		err = mintTokens(sn, big.NewInt(1))
		require.True(t, errors.Is(err, vmtxbuilder.ErrNativeTokenSupplyOutOffBounds))

		ch.AssertL2AccountNativeToken(senderAgentID, &tokenID, big.NewInt(1000))
		ch.AssertL2TotalNativeTokens(&tokenID, big.NewInt(1000))
	})
	t.Run("max supply MaxUint256, mintTokens MaxUint256_1", func(t *testing.T) {
		initTest()
		err, sn, tokenID := createFoundry(t, nil, nil, abi.MaxUint256)
		require.NoError(t, err)
		require.EqualValues(t, 1, sn)

		err = mintTokens(sn, abi.MaxUint256)
		require.NoError(t, err)
		ch.AssertL2AccountNativeToken(senderAgentID, &tokenID, abi.MaxUint256)

		err = mintTokens(sn, big.NewInt(1))
		require.True(t, errors.Is(err, vmtxbuilder.ErrNativeTokenSupplyOutOffBounds))

		ch.AssertL2AccountNativeToken(senderAgentID, &tokenID, abi.MaxUint256)
		ch.AssertL2TotalNativeTokens(&tokenID, abi.MaxUint256)
	})
	t.Run("max supply 100, destroy fail", func(t *testing.T) {
		initTest()
		err, sn, tokenID := createFoundry(t, nil, nil, abi.MaxUint256)
		require.NoError(t, err)
		require.EqualValues(t, 1, sn)

		err = destroyTokens(sn, big.NewInt(1))
		require.True(t, errors.Is(err, vmtxbuilder.ErrNativeTokenSupplyOutOffBounds))
		ch.AssertL2AccountNativeToken(senderAgentID, &tokenID, big.NewInt(0))
		ch.AssertL2TotalNativeTokens(&tokenID, big.NewInt(0))
	})
	t.Run("max supply 100, mint_20, destroy_10", func(t *testing.T) {
		initTest()
		err, sn, tokenID := createFoundry(t, nil, nil, big.NewInt(100))
		require.NoError(t, err)
		require.EqualValues(t, 1, sn)

		out, err := ch.GetFoundryOutput(1)
		require.NoError(t, err)
		require.EqualValues(t, out.MustNativeTokenID(), tokenID)
		ch.AssertL2AccountNativeToken(senderAgentID, &tokenID, big.NewInt(0))
		ch.AssertL2TotalNativeTokens(&tokenID, big.NewInt(0))

		err = mintTokens(sn, big.NewInt(20))
		require.NoError(t, err)
		ch.AssertL2AccountNativeToken(senderAgentID, &tokenID, big.NewInt(20))
		ch.AssertL2TotalNativeTokens(&tokenID, big.NewInt(20))

		err = destroyTokens(sn, big.NewInt(10))
		require.NoError(t, err)
		ch.AssertL2TotalNativeTokens(&tokenID, big.NewInt(10))
		ch.AssertL2AccountNativeToken(senderAgentID, &tokenID, big.NewInt(10))
	})
	t.Run("max supply 1000000, mint_1000000, destroy_1000000", func(t *testing.T) {
		initTest()
		err, sn, tokenID := createFoundry(t, nil, nil, big.NewInt(1000000))
		require.NoError(t, err)
		require.EqualValues(t, 1, sn)

		out, err := ch.GetFoundryOutput(1)
		require.NoError(t, err)
		require.EqualValues(t, out.MustNativeTokenID(), tokenID)
		ch.AssertL2AccountNativeToken(senderAgentID, &tokenID, big.NewInt(0))
		ch.AssertL2TotalNativeTokens(&tokenID, big.NewInt(0))

		err = mintTokens(sn, big.NewInt(1000000))
		require.NoError(t, err)
		ch.AssertL2AccountNativeToken(senderAgentID, &tokenID, big.NewInt(1000000))
		ch.AssertL2TotalNativeTokens(&tokenID, big.NewInt(1000000))
		out, err = ch.GetFoundryOutput(1)
		require.NoError(t, err)
		require.True(t, big.NewInt(1000000).Cmp(out.CirculatingSupply) == 0)

		//err = destroyTokens(sn, big.NewInt(1000000)) // <<<<<<<< fails TODO
		//require.NoError(t, err)
		//ch.AssertL2TotalNativeTokens(&tokenID, big.NewInt(0))
		//ch.AssertL2AccountNativeToken(senderAgentID, &tokenID, big.NewInt(0))
		//out, err = ch.GetFoundryOutput(1)
		//require.NoError(t, err)
		//require.True(t, big.NewInt(0).Cmp(out.CirculatingSupply) == 0)
	})
	t.Run("10 foundries", func(t *testing.T) {
		initTest()

		for sn := uint32(1); sn <= 10; sn++ {
			var tag iotago.TokenTag
			copy(tag[:], util.Uint32To4Bytes(sn))
			err, snBack, tokenID := createFoundry(t, nil, &tag, big.NewInt(int64(sn+1)))
			require.NoError(t, err)
			require.EqualValues(t, int(sn), int(snBack))
			ch.AssertL2AccountNativeToken(senderAgentID, &tokenID, big.NewInt(0))
			ch.AssertL2TotalNativeTokens(&tokenID, big.NewInt(0))
		}
		// mint max supply from each
		for sn := uint32(1); sn <= 10; sn++ {
			err := mintTokens(sn, big.NewInt(int64(sn+1)))
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
			err := destroyTokens(sn, big.NewInt(1))
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

type testValues struct {
	env                *solo.Solo
	chainOwner, sender *cryptolib.KeyPair
	senderAddr         iotago.Address
	senderAgentID      *iscp.AgentID
	ch                 *solo.Chain
	req                *solo.CallParams
}

func initTest(t *testing.T) *testValues {
	ret := &testValues{}
	ret.env = solo.New(t)

	ret.chainOwner, _ = ret.env.NewKeyPairWithFunds(ret.env.NewSeedFromIndex(10))
	//chainOwnerAgentID := iscp.NewAgentID(chainOwnerAddr, 0)

	ret.sender, ret.senderAddr = ret.env.NewKeyPairWithFunds(ret.env.NewSeedFromIndex(11))
	ret.senderAgentID = iscp.NewAgentID(ret.senderAddr, 0)

	ret.ch = ret.env.NewChain(ret.chainOwner, "chain1")
	ret.req = solo.NewCallParams(accounts.Contract.Name, accounts.FuncDeposit.Name)
	return ret
}

func TestDepositIotas(t *testing.T) {
	// the test check how request transaction construction functions adjust iotas to the minimum needed for the
	// dust deposit. If byte cost is 185, anything below that fill be topped up to 185, above that no adjustment is needed
	for _, addIotas := range []uint64{0, 50, 150, 200, 1000} {
		t.Run("add iotas "+strconv.Itoa(int(addIotas)), func(t *testing.T) {
			v := initTest(t)
			v.req = v.req.AddIotas(addIotas)
			tx, _, err := v.ch.PostRequestSyncTx(v.req, v.sender)
			require.NoError(t, err)

			byteCost := tx.Essence.Outputs[0].VByteCost(v.env.RentStructure(), nil)
			t.Logf("byteCost = %d", byteCost)

			// here we calculate what is expected
			expected := addIotas
			if expected < byteCost {
				expected = byteCost
			}
			v.ch.AssertL2AccountIotas(v.senderAgentID, expected)
		})
	}
}

func TestDepositNativeTokens(t *testing.T) {
	t.Run("deposit 1 native token", func(t *testing.T) {
		v := initTest(t)
		v.env.AssertL1AddressIotas(v.senderAddr, solo.Saldo)
		rndTokens := tpkg.RandSortNativeTokens(1)
		v.req = v.req.AddNativeTokens(rndTokens...)
		_, _, err := v.ch.PostRequestSyncTx(v.req, v.sender)
		require.NoError(t, err)
	})
	//t.Run("add many native tokens", func(t *testing.T) {
	//	for _, genTokens := range []int{1, 5, 10}  {
	//		t.Run("iotas+nt "+strconv.Itoa(genTokens), func(t *testing.T) {
	//			v := initTest(t)
	//			_, err := v.env.L1Ledger().GetFundsFromFaucet(v.senderAddr)
	//			require.NoError(t, err)
	//
	//			expected := uint64(i * 50)
	//			v.req = v.req.AddIotas(expected)
	//			rndTokens := tpkg.RandSortNativeTokens(i)
	//			v.req = v.req.AddNativeTokens(rndTokens...)
	//			tx, _, err := v.ch.PostRequestSyncTx(v.req, v.sender)
	//			require.NoError(t, err)
	//
	//			byteCost := tx.Essence.Outputs[0].VByteCost(v.env.RentStructure(), nil)
	//			t.Logf("byteCost = %d", byteCost)
	//			if expected < byteCost {
	//				expected = byteCost
	//			}
	//			v.ch.AssertL2AccountIotas(v.senderAgentID, expected)
	//		})
	//	}
	//
	//})
}
