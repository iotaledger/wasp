package transaction

import (
	"bytes"
	"crypto/ed25519"
	"testing"
	"time"

	"github.com/iotaledger/iota.go/v3/tpkg"

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
	var user cryptolib.KeyPair
	var userAddr, stateAddr *iotago.Ed25519Address
	var err error
	var chainID *iscp.ChainID
	var originTxID *iotago.TransactionID

	initTest := func() {
		u = utxodb.New()
		user, userAddr = u.NewKeyPairByIndex(1)
		_, err := u.RequestFunds(userAddr)
		require.NoError(t, err)

		_, stateAddr = u.NewKeyPairByIndex(2)

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
}

func TestConsumeRequest(t *testing.T) {
	stateController := tpkg.RandEd25519PrivateKey()
	stateControllerAddr := iotago.Ed25519AddressFromPubKey(stateController.Public().(ed25519.PublicKey))
	addrKeys := iotago.AddressKeys{Address: &stateControllerAddr, Keys: stateController}

	aliasOut1 := &iotago.AliasOutput{
		Amount:               1337,
		AliasID:              tpkg.RandAliasAddress().AliasID(),
		StateController:      &stateControllerAddr,
		GovernanceController: &stateControllerAddr,
		StateIndex:           1,
	}
	aliasOut1Inp := tpkg.RandUTXOInput()

	req := &iotago.ExtendedOutput{
		Amount:  1337,
		Address: aliasOut1.AliasID.ToAddress(),
	}
	reqInp := tpkg.RandUTXOInput()

	aliasOut2 := &iotago.AliasOutput{
		Amount:               1337 * 2,
		AliasID:              aliasOut1.AliasID,
		StateController:      &stateControllerAddr,
		GovernanceController: &stateControllerAddr,
		StateIndex:           2,
	}
	essence := &iotago.TransactionEssence{
		Inputs:  iotago.Inputs{aliasOut1Inp, reqInp},
		Outputs: iotago.Outputs{aliasOut2},
	}
	sigs, err := essence.Sign(addrKeys)
	require.NoError(t, err)

	tx := &iotago.Transaction{
		Essence: essence,
		UnlockBlocks: iotago.UnlockBlocks{
			&iotago.SignatureUnlockBlock{Signature: sigs[0]},
			&iotago.AliasUnlockBlock{Reference: 0},
		},
	}
	semValCtx := &iotago.SemanticValidationContext{
		ExtParas: &iotago.ExternalUnlockParameters{
			ConfMsIndex: 1,
			ConfUnix:    uint64(time.Now().Unix()),
		},
	}
	outset := iotago.OutputSet{
		aliasOut1Inp.ID(): aliasOut1,
		reqInp.ID():       req,
	}

	err = tx.SemanticallyValidate(semValCtx, outset)
	require.NoError(t, err)
}
