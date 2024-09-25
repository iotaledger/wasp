package vmtxbuilder

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/sui-go/sui"
)

type TransactionBuilder interface {
	Clone() TransactionBuilder
	ConsumeRequest(req isc.OnLedgerRequest)

	// pt command will be appended into ptb
	SendAssets(target *sui.Address, assets *isc.Assets)
	// pt command will be appended into ptb
	SendCrossChainRequest(targetPackage *sui.Address, targetAnchor *sui.Address, assets *isc.Assets, metadata *isc.SendMetadata)

	// this will reset txb into nil
	BuildTransactionEssence(stateMetadata []byte) sui.ProgrammableTransaction // TODO add stateMetadata?
}
