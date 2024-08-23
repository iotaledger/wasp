package vmtxbuilder

import (
	"math/big"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/sui-go/sui"
)

type TransactionBuilder interface {
	Clone() TransactionBuilder
	ConsumeRequest(req isc.OnLedgerRequest)
	SendObject(object sui.Object) (storageDepositReturned *big.Int)
	BuildTransactionEssence(stateMetadata []byte) sui.ProgrammableTransaction // TODO add stateMetadata?
}
