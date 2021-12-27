package testcore

import (
	"errors"
	"math/big"
	"testing"

	"github.com/iotaledger/wasp/packages/util"

	"github.com/iotaledger/iota.go/v3/tpkg"

	"github.com/iotaledger/wasp/packages/cryptolib"

	"github.com/iotaledger/wasp/packages/kv/kvdecoder"

	"github.com/ethereum/go-ethereum/accounts/abi"

	"github.com/iotaledger/wasp/packages/vm/vmcontext/vmtxbuilder"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
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
		req := solo.NewCallParamsFromDic(accounts.Contract.Name, accounts.FuncFoundryCreateNew.Name, par)
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
		)
		_, _, err := ch.PostRequestSyncTx(req, senderKeyPair)
		return err
	}
	destroyTokens := func(sn uint32, amount *big.Int) error {
		req := solo.NewCallParams(accounts.Contract.Name, accounts.FuncFoundryModifySupply.Name,
			accounts.ParamFoundrySN, sn,
			accounts.ParamSupplyDeltaAbs, amount,
			accounts.ParamDestroyTokens, true,
		)
		_, _, err := ch.PostRequestSyncTx(req, senderKeyPair)
		return err
	}

	t.Run("supply 10", func(t *testing.T) {
		initTest()
		err, sn, _ := createFoundry(t, nil, nil, big.NewInt(10))
		require.EqualValues(t, 1, sn)
		require.NoError(t, err)
	})
	t.Run("supply 1", func(t *testing.T) {
		initTest()
		err, sn, _ := createFoundry(t, nil, nil, big.NewInt(1))
		require.EqualValues(t, 1, sn)
		require.NoError(t, err)
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
		require.EqualValues(t, 1, sn)
		require.NoError(t, err)
	})
	t.Run("supply max possible", func(t *testing.T) {
		initTest()
		err, sn, _ := createFoundry(t, nil, nil, abi.MaxUint256)
		require.EqualValues(t, 1, sn)
		require.NoError(t, err)
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
		require.EqualValues(t, 1, sn)
		require.NoError(t, err)
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
		require.EqualValues(t, 1, sn)
		require.NoError(t, err)
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
		require.EqualValues(t, 1, sn)
		require.NoError(t, err)

		err = mintTokens(sn, big.NewInt(2))
		require.True(t, errors.Is(err, vmtxbuilder.ErrNativeTokenSupplyOutOffBounds))

		ch.AssertL2AccountNativeToken(senderAgentID, &tokenID, big.NewInt(0))
		ch.AssertL2TotalNativeTokens(&tokenID, big.NewInt(0))
	})
	t.Run("max supply 1000, mintTokens 500_500_1", func(t *testing.T) {
		initTest()
		err, sn, tokenID := createFoundry(t, nil, nil, big.NewInt(1000))
		require.EqualValues(t, 1, sn)
		require.NoError(t, err)

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
		require.EqualValues(t, 1, sn)
		require.NoError(t, err)

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
		require.EqualValues(t, 1, sn)
		require.NoError(t, err)

		err = destroyTokens(sn, big.NewInt(1))
		require.True(t, errors.Is(err, vmtxbuilder.ErrNativeTokenSupplyOutOffBounds))
		ch.AssertL2AccountNativeToken(senderAgentID, &tokenID, big.NewInt(0))
		ch.AssertL2TotalNativeTokens(&tokenID, big.NewInt(0))
	})
	t.Run("max supply 100, mint_20, destroy_10", func(t *testing.T) {
		initTest()
		err, sn, tokenID := createFoundry(t, nil, nil, big.NewInt(100))
		require.EqualValues(t, 1, sn)
		require.NoError(t, err)

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
		require.EqualValues(t, 1, sn)
		require.NoError(t, err)

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

		err = destroyTokens(sn, big.NewInt(1000000)) // <<<<<<<< fails TODO
		require.NoError(t, err)
		ch.AssertL2TotalNativeTokens(&tokenID, big.NewInt(0))
		ch.AssertL2AccountNativeToken(senderAgentID, &tokenID, big.NewInt(0))
		out, err = ch.GetFoundryOutput(1)
		require.NoError(t, err)
		require.True(t, big.NewInt(0).Cmp(out.CirculatingSupply) == 0)
	})
	t.Run("10 foundries", func(t *testing.T) {
		initTest()

		for sn := uint32(1); sn <= 10; sn++ {
			var tag iotago.TokenTag
			copy(tag[:], util.Uint32To4Bytes(sn))
			err, snBack, tokenID := createFoundry(t, nil, &tag, big.NewInt(int64(sn+1)))
			require.EqualValues(t, int(sn), int(snBack))
			require.NoError(t, err)
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
		require.NoError(t, err)
	})
	t.Run("pass", func(t *testing.T) {
		err := iotago.NativeTokenSumBalancedWithDiff(tokenID, inSums, outSumsGood, circSupplyChange)
		require.NoError(t, err)
	})
}
