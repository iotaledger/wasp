package state

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/variables"
	"io"
)

type VariableState interface {
	StateIndex() uint32
	Apply(StateUpdate) VariableState
	SaveToDb() error
	Variables() variables.Variables
	Read(io.Reader) error // TODO serialization must be replaced for more efficient mechanism
	Write(io.Writer) error
}

// state update with anchor transaction hash
// NOTE: if error occures during processing, it is a deterministic result
// special variable "ErrorMsg" in stateUpdate(and variable state) must be set to error string,
// otherwise it is a normal state update
type StateUpdate interface {
	Address() *address.Address
	StateIndex() uint32
	StateTransactionId() valuetransaction.ID
	SetStateTransactionId(valuetransaction.ID)
	SaveToDb() error
	Variables() variables.Variables
	Read(io.Reader) error
	Write(io.Writer) error
}
