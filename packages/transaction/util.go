package transaction

import (
	"errors"

	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
)

var ErrNoAliasOutputAtIndex0 = errors.New("origin AliasOutput not found at index 0")

// GetAnchorFromTransaction analyzes the output at index 0 and extracts anchor information. Otherwise error
func GetAnchorFromTransaction(tx *iotago.Transaction) (*isc.StateAnchor, *iotago.AliasOutput, error) {
	panic("refactor me: probabyly remove GetAnchorFromTransaction as we don't need to filter the anchor out of a TX anymore")
	/*anchorOutput, ok := tx.Essence.Outputs[0].(*iotago.AliasOutput)
	if !ok {
		return nil, nil, ErrNoAliasOutputAtIndex0
	}
	txid, err := tx.ID()
	if err != nil {
		return nil, anchorOutput, fmt.Errorf("GetAnchorFromTransaction: %w", err)
	}
	aliasID := anchorOutput.AliasID
	isOrigin := false

	if aliasID.Empty() {
		isOrigin = true
		aliasID = iotago.AliasIDFromOutputID(iotago.OutputIDFromTransactionIDAndIndex(txid, 0))
	}
	return &isc.StateAnchor{
		IsOrigin:             isOrigin,
		OutputID:             iotago.OutputIDFromTransactionIDAndIndex(txid, 0),
		ChainID:              isc.ChainIDFromAliasID(aliasID),
		StateController:      lo.Must(cryptolib.NewAddressFromHexString(anchorOutput.StateController().String())),
		GovernanceController: lo.Must(cryptolib.NewAddressFromHexString(anchorOutput.GovernorAddress().String())),
		StateIndex:           anchorOutput.StateIndex,
		StateData:            anchorOutput.StateMetadata,
		Deposit:              anchorOutput.Amount,
	}, anchorOutput, nil*/
}

func MakeSignatureAndReferenceUnlocks(totalInputs int, sig *cryptolib.Signature) iotago.Unlocks {
	ret := make(iotago.Unlocks, totalInputs)
	for i := range ret {
		if i == 0 {
			panic("refactor me: AsIotaSignature")

			// ret[0] = &iotago.SignatureUnlock{Signature: sig.AsSuiSignature()} // TODO: move SignatureUnlock to isc-private?
			continue
		}
		ret[i] = &iotago.ReferenceUnlock{Reference: 0}
	}
	return ret
}

func MakeSignatureAndAliasUnlockFeatures(totalInputs int, sig *cryptolib.Signature) iotago.Unlocks {
	ret := make(iotago.Unlocks, totalInputs)
	for i := range ret {
		if i == 0 {
			panic("refactor me: AsIotaSignature")
			// ret[0] = &iotago.SignatureUnlock{Signature: sig.AsSuiSignature()} // TODO: move SignatureUnlock to isc-private?
			continue
		}
		ret[i] = &iotago.AliasUnlock{Reference: 0}
	}
	return ret
}

func MakeAnchorTransaction(essence *iotago.TransactionEssence, sig *cryptolib.Signature) *iotago.Transaction {
	return &iotago.Transaction{
		Essence: essence,
		Unlocks: MakeSignatureAndAliasUnlockFeatures(len(essence.Inputs), sig),
	}
}

func CreateAndSignTx(inputs iotago.Inputs, inputsCommitment []byte, outputs iotago.Outputs, wallet cryptolib.Signer, networkID uint64) (*iotago.Transaction, error) {
	unorderedEssence := &iotago.TransactionEssence{
		NetworkID: networkID,
		Inputs:    inputs,
		Outputs:   outputs,
	}

	// IMPORTANT: serialize and de-serialize the essence, just to make sure it is correctly ordered before signing
	// otherwise it might fail when it reaches the node, since the PoW that would order the tx is done after the signing,
	// so if we don't order now, we might sign an invalid TX
	essenceBytes, err := unorderedEssence.Serialize(serializer.DeSeriModePerformValidation|serializer.DeSeriModePerformLexicalOrdering, nil)
	if err != nil {
		return nil, err
	}
	essence := new(iotago.TransactionEssence)
	_, err = essence.Deserialize(essenceBytes, serializer.DeSeriModeNoValidation, nil)
	if err != nil {
		return nil, err
	}
	// --

	sigs, err := SignEssence(essence, inputsCommitment, wallet)
	if err != nil {
		return nil, err
	}

	return &iotago.Transaction{
		Essence: essence,
		Unlocks: MakeSignatureAndReferenceUnlocks(len(inputs), sigs[0]),
	}, nil
}
