package vmtxbuilder

import (
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/packages/isc"
)

// TODO maybe we dont need this interface
type TransactionBuilder interface {
	Clone() TransactionBuilder
	ConsumeRequest(req isc.OnLedgerRequest)

	// pt command will be appended into ptb
	SendAssets(target *iotago.Address, assets *isc.Assets)
	// pt command will be appended into ptb
	SendCrossChainRequest(targetPackage *iotago.Address, targetAnchor *iotago.Address, assets *isc.Assets, metadata *isc.SendMetadata)

	// this will be appended RotationTransaction PTB in the end of 'BuildTransactionEssence()'
	RotationTransaction(rotationAddress *iotago.Address)

	// this will reset txb into nil
	BuildTransactionEssence(stateMetadata []byte, topUpAmount uint64) iotago.ProgrammableTransaction

	// return a clone of PTB object for viewing
	ViewPTB() *iotago.ProgrammableTransactionBuilder

	// CheckTransactionSize returns ErrMaxTransactionSizeExceeded if the estimated transaction size exceeds the limit
	CheckTransactionSize() error
}
