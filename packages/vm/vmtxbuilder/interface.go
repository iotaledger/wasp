package vmtxbuilder

import (
	sui2 "github.com/iotaledger/wasp/clients/iota-go/sui"
	"github.com/iotaledger/wasp/packages/isc"
)

type TransactionBuilder interface {
	Clone() TransactionBuilder
	ConsumeRequest(req isc.OnLedgerRequest)

	// pt command will be appended into ptb
	SendAssets(target *sui2.Address, assets *isc.Assets)
	// pt command will be appended into ptb
	SendCrossChainRequest(targetPackage *sui2.Address, targetAnchor *sui2.Address, assets *isc.Assets, metadata *isc.SendMetadata)

	// this will reset txb into nil
	BuildTransactionEssence(stateMetadata []byte) sui2.ProgrammableTransaction // TODO add stateMetadata?
}
