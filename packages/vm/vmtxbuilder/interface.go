package vmtxbuilder

import (
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/packages/isc"
)

type TransactionBuilder interface {
	Clone() TransactionBuilder
	ConsumeRequest(req isc.OnLedgerRequest)

	// pt command will be appended into ptb
	SendAssets(target *iotago.Address, assets *isc.Assets)
	// pt command will be appended into ptb
	SendCrossChainRequest(targetPackage *iotago.Address, targetAnchor *iotago.Address, assets *isc.Assets, metadata *isc.SendMetadata)

	// this will reset txb into nil
	BuildTransactionEssence(stateMetadata []byte) iotago.ProgrammableTransaction // TODO add stateMetadata?
}
