package vmtxbuilder

import (
	"math/big"
	"testing"

	iotago "github.com/iotaledger/iota.go/v3"
)

func TestNewTxBuilder(t *testing.T) {
	anchor := &iotago.AliasOutput{
		Amount:               0,
		NativeTokens:         nil,
		AliasID:              iotago.AliasID{},
		StateController:      nil,
		GovernanceController: nil,
		StateIndex:           0,
		StateMetadata:        nil,
		FoundryCounter:       0,
		Blocks:               nil,
	}
	anchorID := iotago.UTXOInput{
		TransactionID:          iotago.TransactionID{},
		TransactionOutputIndex: 0,
	}
	_ = NewAnchorTransactionBuilder(anchor, anchorID, 0, func(id iotago.NativeTokenID) (*big.Int, iotago.UTXOInput) {
		return nil, iotago.UTXOInput{}
	})
}
