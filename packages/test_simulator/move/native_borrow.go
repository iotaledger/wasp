package move

import (
	"fmt"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/packages/test_simulator/l1"
)

var BorrowHandlers = map[string]l1.MoveCallFunc{
	"new":      borrowNew,
	"borrow":   borrowBorrow,
	"put_back": borrowPutBack,
	"destroy":  borrowDestroy,
}

// ReferentValue represents a Referent<T> at runtime.
type ReferentValue struct {
	ID    iotago.ObjectID
	Value *l1.Value // nil when borrowed out
}

// borrow::new(inner, ctx) -> Referent<T>
func borrowNew(ctx *l1.CallContext, _ *iotago.ProgrammableMoveCall, args []l1.Value) ([]l1.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("borrow::new requires 1 argument")
	}
	id := ctx.FreshID()
	return []l1.Value{{
		ObjectID: &id,
		Raw: &ReferentValue{
			ID:    id,
			Value: &args[0],
		},
		Type: "Referent",
	}}, nil
}

// borrow::borrow(referent) -> (inner, Borrow)
func borrowBorrow(_ *l1.CallContext, _ *iotago.ProgrammableMoveCall, args []l1.Value) ([]l1.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("borrow::borrow requires 1 argument")
	}
	ref, ok := args[0].Raw.(*ReferentValue)
	if !ok {
		return nil, fmt.Errorf("borrow::borrow: argument is not a Referent")
	}
	if ref.Value == nil {
		return nil, fmt.Errorf("borrow::borrow: referent already borrowed")
	}

	inner := *ref.Value
	token := &l1.BorrowToken{
		ReferentID: ref.ID,
		RefAddr:    ref.ID,
	}
	ref.Value = nil

	return []l1.Value{inner, {Raw: token, Type: "Borrow"}}, nil
}

// borrow::put_back(referent, inner, borrow_token)
func borrowPutBack(_ *l1.CallContext, _ *iotago.ProgrammableMoveCall, args []l1.Value) ([]l1.Value, error) {
	if len(args) < 3 {
		return nil, fmt.Errorf("borrow::put_back requires 3 arguments")
	}
	ref, ok := args[0].Raw.(*ReferentValue)
	if !ok {
		return nil, fmt.Errorf("borrow::put_back: first argument is not a Referent")
	}
	token, ok := args[2].Raw.(*l1.BorrowToken)
	if !ok {
		return nil, fmt.Errorf("borrow::put_back: third argument is not a BorrowToken")
	}
	if token.ReferentID != ref.ID {
		return nil, fmt.Errorf("borrow::put_back: borrow token does not match referent")
	}

	ref.Value = &args[1]
	return nil, nil
}

// borrow::destroy(referent) -> inner
func borrowDestroy(_ *l1.CallContext, _ *iotago.ProgrammableMoveCall, args []l1.Value) ([]l1.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("borrow::destroy requires 1 argument")
	}
	ref, ok := args[0].Raw.(*ReferentValue)
	if !ok {
		return nil, fmt.Errorf("borrow::destroy: argument is not a Referent")
	}
	if ref.Value == nil {
		return nil, fmt.Errorf("borrow::destroy: referent is empty (still borrowed)")
	}
	return []l1.Value{*ref.Value}, nil
}
