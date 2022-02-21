package transaction

import (
	"bytes"
	"testing"
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/utxodb"
	"github.com/stretchr/testify/require"
)

func TestCreateOrigin(t *testing.T) {
	var u *utxodb.UtxoDB
	var originTx, txInit *iotago.Transaction
	var user *cryptolib.KeyPair
	var userAddr, stateAddr *iotago.Ed25519Address
	var err error
	var chainID *iscp.ChainID
	var originTxID *iotago.TransactionID

	initTest := func() {
		u = utxodb.New()
		user, userAddr = u.NewKeyPairByIndex(1)
		_, err := u.GetFundsFromFaucet(userAddr)
		require.NoError(t, err)

		_, stateAddr = u.NewKeyPairByIndex(2)

		require.EqualValues(t, utxodb.FundsFromFaucetAmount, u.GetAddressBalanceIotas(userAddr))
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
			parameters.L1ForTesting(),
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
			parameters.L1ForTesting(),
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

		require.EqualValues(t, utxodb.FundsFromFaucetAmount-chainIotas-initIotas, int(u.GetAddressBalanceIotas(userAddr)))
		require.EqualValues(t, 0, u.GetAddressBalanceIotas(stateAddr))
		allOutputs, ids := u.GetUnspentOutputs(chainID.AsAddress())
		require.EqualValues(t, 2, len(allOutputs))
		require.EqualValues(t, 2, len(ids))
	})
}

func TestConsumeRequest(t *testing.T) {
	stateControllerKeyPair := cryptolib.NewKeyPair()
	stateController := stateControllerKeyPair.GetPrivateKey()
	stateControllerAddr := stateControllerKeyPair.GetPublicKey().AsEd25519Address()
	addrKeys := stateController.AddressKeysForEd25519Address(stateControllerAddr)

	aliasOut1ID := tpkg.RandOutputID(0)
	aliasOut1 := &iotago.AliasOutput{
		Amount:     1337,
		AliasID:    tpkg.RandAliasAddress().AliasID(),
		StateIndex: 1,
		Conditions: iotago.UnlockConditions{
			&iotago.StateControllerAddressUnlockCondition{Address: stateControllerAddr},
			&iotago.GovernorAddressUnlockCondition{Address: stateControllerAddr},
		},
	}
	aliasOut1Inp := tpkg.RandUTXOInput()

	reqID := tpkg.RandOutputID(1)
	req := &iotago.BasicOutput{
		Amount: 1337,
		Conditions: iotago.UnlockConditions{
			&iotago.AddressUnlockCondition{Address: aliasOut1.AliasID.ToAddress()},
		},
	}
	reqInp := tpkg.RandUTXOInput()

	aliasOut2 := &iotago.AliasOutput{
		Amount:     1337 * 2,
		AliasID:    aliasOut1.AliasID,
		StateIndex: 2,
		Conditions: iotago.UnlockConditions{
			&iotago.StateControllerAddressUnlockCondition{Address: stateControllerAddr},
			&iotago.GovernorAddressUnlockCondition{Address: stateControllerAddr},
		},
	}
	essence := &iotago.TransactionEssence{
		NetworkID: 0,
		Inputs:    iotago.Inputs{aliasOut1Inp, reqInp},
		Outputs:   iotago.Outputs{aliasOut2},
	}
	sigs, err := essence.Sign(
		iotago.OutputIDs{aliasOut1ID, reqID}.
			OrderedSet(iotago.OutputSet{aliasOut1ID: aliasOut1, reqID: req}).
			MustCommitment(),
		addrKeys,
	)
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
			ConfUnix:    uint32(time.Now().Unix()),
		},
	}
	outset := iotago.OutputSet{
		aliasOut1Inp.ID(): aliasOut1,
		reqInp.ID():       req,
	}

	err = tx.SemanticallyValidate(semValCtx, outset)
	require.NoError(t, err)
}
