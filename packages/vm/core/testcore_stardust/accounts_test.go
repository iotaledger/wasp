package testcore

import (
	"errors"
	"math/big"
	"testing"

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

func TestCreateFoundry(t *testing.T) {
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
	mint := func(sn uint32, amount *big.Int) error {
		req := solo.NewCallParams(accounts.Contract.Name, accounts.FuncFoundryModifySupply.Name,
			accounts.ParamFoundrySN, sn,
			accounts.ParamSupplyDeltaAbs, amount,
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

	t.Run("max supply 10, mint 5", func(t *testing.T) {
		initTest()
		err, sn, tokenID := createFoundry(t, nil, nil, big.NewInt(10))
		require.EqualValues(t, 1, sn)
		require.NoError(t, err)

		err = mint(sn, big.NewInt(5))
		require.NoError(t, err)

		ch.AssertL2AccountNativeToken(senderAgentID, &tokenID, big.NewInt(5))
	})
	t.Run("max supply 1, mint 1", func(t *testing.T) {
		initTest()
		err, sn, tokenID := createFoundry(t, nil, nil, big.NewInt(1))
		require.EqualValues(t, 1, sn)
		require.NoError(t, err)

		err = mint(sn, big.NewInt(1))
		require.NoError(t, err)

		ch.AssertL2AccountNativeToken(senderAgentID, &tokenID, big.NewInt(1))
	})

	t.Run("max supply 1, mint 2", func(t *testing.T) {
		initTest()
		err, sn, tokenID := createFoundry(t, nil, nil, big.NewInt(1))
		require.EqualValues(t, 1, sn)
		require.NoError(t, err)

		err = mint(sn, big.NewInt(2))
		require.True(t, errors.Is(err, vmtxbuilder.ErrNativeTokenSupplyOutOffBounds))

		ch.AssertL2AccountNativeToken(senderAgentID, &tokenID, big.NewInt(0))
	})

}
