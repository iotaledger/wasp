package vmtxbuilder

import (
	"math/big"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/sui-go/sui"
)

type TransactionBuilder interface {
	Clone() TransactionBuilder
	ConsumeRequest(req isc.OnLedgerRequest) (storageDepositNeeded *big.Int)
	SendObject(object sui.Object) (storageDepositReturned *big.Int)

	// TODO there a "dry-run" mode, but that is "l1 server-side". Is there away to verifyt his without an external request?
	MustBalanced()

	// TODO
	// - add stateMetadata?
	// - what exactly needs to be signed by the committee
	BuildTransactionEssence(stateRoot *state.L1Commitment) (msg []byte)
}
