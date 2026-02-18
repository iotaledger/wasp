package l1

import "github.com/iotaledger/wasp/v2/clients/iota-go/iotago"

// Value represents a runtime value passed between PTB commands.
// It can hold any type: object references, balances, receipts, vectors, etc.
type Value struct {
	ObjectID *iotago.ObjectID // if this value references an object in the store
	Raw      any              // the actual Go-level value (balance amount, borrow token, etc.)
	Type     string           // Move type string for tracking
}

// BorrowToken is the HotPotato token for borrow::Borrow.
type BorrowToken struct {
	ReferentID iotago.ObjectID
	RefAddr    iotago.ObjectID // the Referent's address (for put_back validation)
}

// BalanceValue represents a Move Balance<T> value.
type BalanceValue struct {
	CoinType string
	Amount   uint64
}

// CallContext provides access to the execution state during a MoveCall.
type CallContext struct {
	Store     *ObjectStore
	Sender    iotago.Address
	TxDigest  iotago.TransactionDigest
	IDCounter *uint64
	PackageID iotago.PackageID
}

// FreshID generates a new ObjectID.
func (ctx *CallContext) FreshID() iotago.ObjectID {
	return FreshID(ctx.TxDigest, ctx.IDCounter)
}

// MoveCallFunc is the standard handler signature for a single Move function.
type MoveCallFunc func(ctx *CallContext, call *iotago.ProgrammableMoveCall, args []Value) ([]Value, error)

// MoveCallHandler processes MoveCall commands.
type MoveCallHandler interface {
	ExecuteMoveCall(ctx *CallContext, call *iotago.ProgrammableMoveCall, args []Value) ([]Value, error)
}
