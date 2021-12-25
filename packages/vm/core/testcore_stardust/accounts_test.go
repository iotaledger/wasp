package testcore

import (
	"errors"
	"math/big"
	"testing"

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
	createFoundry := func(sch *iotago.TokenScheme, tag *iotago.TokenTag, maxSupply *big.Int) error {
		env := solo.New(t)
		env.EnablePublisher(true)
		ch := env.NewChain(nil, "chain1")
		defer env.WaitPublisher()
		defer ch.Log.Sync()

		senderKeyPair, senderAddr := env.NewKeyPairWithFunds(env.NewSeedFromIndex(10))
		_ = iscp.NewAgentID(senderAddr, 0)

		par := dict.New()
		if sch != nil {
			par.Set(accounts.ParamsTokenScheme, codec.EncodeTokenScheme(*sch))
		}
		if tag != nil {
			par.Set(accounts.ParamsTokenTag, codec.EncodeTokenTag(*tag))
		}
		if maxSupply != nil {
			par.Set(accounts.ParamsMaxSupply, codec.EncodeBigIntAbs(maxSupply))
		}
		req := solo.NewCallParamsFromDic(accounts.Contract.Name, accounts.FuncFoundryCreateNew.Name, par)
		_, _, err := ch.PostRequestSyncTx(req, senderKeyPair)

		return err
	}

	t.Run("supply 10", func(t *testing.T) {
		err := createFoundry(nil, nil, big.NewInt(10))
		require.NoError(t, err)
	})
	t.Run("supply 1", func(t *testing.T) {
		err := createFoundry(nil, nil, big.NewInt(1))
		require.NoError(t, err)
	})
	t.Run("supply 0", func(t *testing.T) {
		err := createFoundry(nil, nil, big.NewInt(0))
		require.True(t, errors.Is(err, vmtxbuilder.ErrCreateFoundryMaxSupplyMustBePositive))
	})
	t.Run("supply negative", func(t *testing.T) {
		err := createFoundry(nil, nil, big.NewInt(-1))
		// encoding will ignore sign
		require.NoError(t, err)
	})
	t.Run("supply max possible", func(t *testing.T) {
		err := createFoundry(nil, nil, abi.MaxUint256)
		require.NoError(t, err)
	})
	t.Run("supply exceed max possible", func(t *testing.T) {
		maxSupply := big.NewInt(0).Set(abi.MaxUint256)
		maxSupply.Add(maxSupply, big.NewInt(1))
		err := createFoundry(nil, nil, maxSupply)
		require.True(t, errors.Is(err, vmtxbuilder.ErrCreateFoundryMaxSupplyTooBig))
	})
}
