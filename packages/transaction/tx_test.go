package transaction

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/utxodb"
)

func TestCreateOrigin(t *testing.T) {
	var u *utxodb.UtxoDB
	var originTx, txInit *iotago.Transaction
	var userKey *cryptolib.KeyPair
	var userAddr, stateAddr *iotago.Ed25519Address
	var err error
	var chainID *isc.ChainID
	var originTxID iotago.TransactionID

	initTest := func() {
		u = utxodb.New()
		userKey = cryptolib.NewKeyPair()
		userAddr = userKey.GetPublicKey().AsEd25519Address()
		_, err := u.GetFundsFromFaucet(userAddr)
		require.NoError(t, err)

		stateKey := cryptolib.NewKeyPair()
		stateAddr = stateKey.GetPublicKey().AsEd25519Address()

		require.EqualValues(t, utxodb.FundsFromFaucetAmount, u.GetAddressBalanceBaseTokens(userAddr))
		require.EqualValues(t, 0, u.GetAddressBalanceBaseTokens(stateAddr))
	}
	createOrigin := func() {
		allOutputs, ids := u.GetUnspentOutputs(userAddr)

		originTx, chainID, err = NewChainOriginTransaction(
			userKey,
			stateAddr,
			stateAddr,
			1000,
			allOutputs,
			ids,
		)
		require.NoError(t, err)

		err = u.AddToLedger(originTx)
		require.NoError(t, err)

		originTxID, err = originTx.ID()
		require.NoError(t, err)

		txBack, ok := u.GetTransaction(originTxID)
		require.True(t, ok)
		txidBack, err := txBack.ID()
		require.NoError(t, err)
		require.EqualValues(t, originTxID, txidBack)

		t.Logf("New chain ID: %s", chainID.String())
	}
	createInitChainTx := func() {
		allOutputs, ids := u.GetUnspentOutputs(userAddr)
		txInit, err = NewRootInitRequestTransaction(
			userKey,
			chainID,
			"test chain",
			allOutputs,
			ids,
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
		require.True(t, bytes.Equal(state.OriginL1Commitment().Bytes(), anchor.StateData))

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

		chainBaseTokens := originTx.Essence.Outputs[0].Deposit()
		initBaseTokens := txInit.Essence.Outputs[0].Deposit()

		t.Logf("chainBaseTokens: %d initBaseTokens: %d", chainBaseTokens, initBaseTokens)

		require.EqualValues(t, utxodb.FundsFromFaucetAmount-chainBaseTokens-initBaseTokens, int(u.GetAddressBalanceBaseTokens(userAddr)))
		require.EqualValues(t, 0, u.GetAddressBalanceBaseTokens(stateAddr))
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

	aliasOutput1ID := tpkg.RandOutputID(0)
	aliasOutput1 := &iotago.AliasOutput{
		Amount:     1337,
		AliasID:    tpkg.RandAliasAddress().AliasID(),
		StateIndex: 1,
		Conditions: iotago.UnlockConditions{
			&iotago.StateControllerAddressUnlockCondition{Address: stateControllerAddr},
			&iotago.GovernorAddressUnlockCondition{Address: stateControllerAddr},
		},
	}
	aliasOutput1UTXOInput := tpkg.RandUTXOInput()

	reqID := tpkg.RandOutputID(1)
	request := &iotago.BasicOutput{
		Amount: 1337,
		Conditions: iotago.UnlockConditions{
			&iotago.AddressUnlockCondition{Address: aliasOutput1.AliasID.ToAddress()},
		},
	}
	requestUTXOInput := tpkg.RandUTXOInput()

	aliasOut2 := &iotago.AliasOutput{
		Amount:     1337 * 2,
		AliasID:    aliasOutput1.AliasID,
		StateIndex: 2,
		Conditions: iotago.UnlockConditions{
			&iotago.StateControllerAddressUnlockCondition{Address: stateControllerAddr},
			&iotago.GovernorAddressUnlockCondition{Address: stateControllerAddr},
		},
	}
	essence := &iotago.TransactionEssence{
		NetworkID: tpkg.TestNetworkID,
		Inputs:    iotago.Inputs{aliasOutput1UTXOInput, requestUTXOInput},
		Outputs:   iotago.Outputs{aliasOut2},
	}
	sigs, err := essence.Sign(
		iotago.OutputIDs{aliasOutput1ID, reqID}.
			OrderedSet(iotago.OutputSet{aliasOutput1ID: aliasOutput1, reqID: request}).
			MustCommitment(),
		addrKeys,
	)
	require.NoError(t, err)

	tx := &iotago.Transaction{
		Essence: essence,
		Unlocks: iotago.Unlocks{
			&iotago.SignatureUnlock{Signature: sigs[0]},
			&iotago.AliasUnlock{Reference: 0},
		},
	}
	semValCtx := &iotago.SemanticValidationContext{
		ExtParas: &iotago.ExternalUnlockParameters{
			ConfUnix: uint32(time.Now().Unix()),
		},
	}
	outset := iotago.OutputSet{
		aliasOutput1UTXOInput.ID(): aliasOutput1,
		requestUTXOInput.ID():      request,
	}

	err = tx.SemanticallyValidate(semValCtx, outset)
	require.NoError(t, err)
}
