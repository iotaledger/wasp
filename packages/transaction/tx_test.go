package transaction

import (
	"bytes"
	"testing"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/iscp"

	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/testutil/testdeserparams"
	"github.com/iotaledger/wasp/packages/utxodb"
	"github.com/stretchr/testify/require"
)

func TestCreateOrigin(t *testing.T) {
	var u *utxodb.UtxoDB
	var originTx, txInit *iotago.Transaction
	var user, stateControllerKeyPair cryptolib.KeyPair
	var userAddr, stateAddr *iotago.Ed25519Address
	var err error
	var chainID *iscp.ChainID
	var originTxID, initTxID *iotago.TransactionID

	initTest := func() {
		u = utxodb.New()
		user, userAddr = u.NewKeyPairByIndex(1)
		_, err := u.RequestFunds(userAddr)
		require.NoError(t, err)

		stateControllerKeyPair, stateAddr = u.NewKeyPairByIndex(2)

		require.EqualValues(t, utxodb.RequestFundsAmount, u.GetAddressBalanceIotas(userAddr))
		require.EqualValues(t, 0, u.GetAddressBalanceIotas(stateAddr))
	}
	createOrigin := func() {
		allOutputs, ids := u.GetUnspentOutputs(userAddr)

		originTx, chainID, err = NewChainOriginTransaction(
			user,
			stateAddr,
			stateAddr,
			0,
			allOutputs,
			ids,
			testdeserparams.DeSerializationParameters(),
		)
		require.NoError(t, err)

		err = u.AddToLedger(originTx)
		require.NoError(t, err)

		originTxID, err = originTx.ID()
		require.NoError(t, err)

		txBack, ok := u.GetTransaction(*originTxID)
		require.True(t, ok)
		txidBack, err := txBack.ID()
		require.NoError(t, err)
		require.EqualValues(t, *originTxID, *txidBack)

		t.Logf("New chain ID: %s", chainID.String())
	}
	createInitChainTx := func() {
		allOutputs, ids := u.GetUnspentOutputs(userAddr)
		txInit, err = NewRootInitRequestTransaction(
			user,
			chainID,
			"test chain",
			allOutputs,
			ids,
			testdeserparams.DeSerializationParameters(),
		)
		require.NoError(t, err)

		initTxID, err = txInit.ID()
		err = u.AddToLedger(txInit)
		require.NoError(t, err)
	}

	t.Run("create origin", func(t *testing.T) {
		initTest()
		createOrigin()

		anchor, _, err := GetAnchorFromTransaction(originTx)
		require.NoError(t, err)
		require.True(t, anchor.IsOrigin)
		require.EqualValues(t, *chainID, anchor.ChainID)
		require.EqualValues(t, 0, anchor.StateIndex)
		require.True(t, stateAddr.Equal(anchor.StateController))
		require.True(t, stateAddr.Equal(anchor.GovernanceController))
		require.True(t, bytes.Equal(state.OriginStateHash().Bytes(), anchor.StateData.Bytes()))

		// only one output is expected in the ledger under the address of chainID
		outs, ids := u.GetUnspentOutputs(chainID.AsAddress())
		require.EqualValues(t, 1, len(outs))
		require.EqualValues(t, 1, len(ids))

		out := u.GetOutput(anchor.OutputID)
		require.NotNil(t, out)
	})
	t.Run("create init chain originTx", func(t *testing.T) {
		initTest()
		createOrigin()
		createInitChainTx()

		chainIotas := originTx.Essence.Outputs[0].Deposit()
		initIotas := txInit.Essence.Outputs[0].Deposit()

		t.Logf("chainIotas: %d initIotas: %d", chainIotas, initIotas)

		require.EqualValues(t, utxodb.RequestFundsAmount-chainIotas-initIotas, int(u.GetAddressBalanceIotas(userAddr)))
		require.EqualValues(t, 0, u.GetAddressBalanceIotas(stateAddr))
		allOutputs, ids := u.GetUnspentOutputs(chainID.AsAddress())
		require.EqualValues(t, 2, len(allOutputs))
		require.EqualValues(t, 2, len(ids))
	})
	t.Run("state transition 0->1", func(t *testing.T) {
		// typical use case in ISC VM: alias consumes ExtendedOutput as request
		initTest()
		createOrigin()
		createInitChainTx()

		txb := iotago.NewTransactionBuilder()
		txb.AddInput(&iotago.ToBeSignedUTXOInput{
			Address: stateAddr,
			Input: &iotago.UTXOInput{
				TransactionID:          *originTxID,
				TransactionOutputIndex: 0,
			},
		})
		txb.AddInput(&iotago.ToBeSignedUTXOInput{
			Address: stateAddr,
			Input: &iotago.UTXOInput{
				TransactionID:          *initTxID,
				TransactionOutputIndex: 0,
			},
		})
		chainUTXO := originTx.Essence.Outputs[0].(*iotago.AliasOutput)
		initRequestUTXO := txInit.Essence.Outputs[0]
		out := &iotago.AliasOutput{
			Amount:               chainUTXO.Deposit() + initRequestUTXO.Deposit(),
			NativeTokens:         nil,
			AliasID:              iotago.AliasID(*chainID),
			StateController:      chainUTXO.StateController,
			GovernanceController: chainUTXO.GovernanceController,
			StateIndex:           1,
			StateMetadata:        nil,
			FoundryCounter:       0,
			Blocks: iotago.FeatureBlocks{
				&iotago.SenderFeatureBlock{
					Address: chainID.AsAddress(),
				},
			},
		}
		txb.AddOutput(out)
		signer := iotago.NewInMemoryAddressSigner(iotago.NewAddressKeysForEd25519Address(stateAddr, stateControllerKeyPair.PrivateKey))
		txNext, err := txb.Build(testdeserparams.DeSerializationParameters(), signer)
		require.NoError(t, err)

		err = u.AddToLedger(txNext)
		require.NoError(t, err)
	})
}
