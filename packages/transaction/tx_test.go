package transaction_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/iotaledger/wasp/packages/cryptolib"
)

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
