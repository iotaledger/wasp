package move

import (
	"fmt"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/packages/test_simulator/l1"
)

var OptionHandlers = map[string]l1.MoveCallFunc{
	"some":         optionSome,
	"none":         optionNone,
	"destroy_some": optionDestroySome,
	"destroy_none": optionDestroyNone,
	"is_some":      optionIsSome,
}

// OptionValue represents an Option<T> at runtime.
type OptionValue struct {
	IsSome bool
	Value  *l1.Value
}

// option::some<T>(value: T) -> Option<T>
func optionSome(_ *l1.CallContext, _ *iotago.ProgrammableMoveCall, args []l1.Value) ([]l1.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("option::some requires 1 argument")
	}
	return []l1.Value{{Raw: &OptionValue{IsSome: true, Value: &args[0]}, Type: "Option"}}, nil
}

// option::none<T>() -> Option<T>
func optionNone(_ *l1.CallContext, _ *iotago.ProgrammableMoveCall, _ []l1.Value) ([]l1.Value, error) {
	return []l1.Value{{Raw: &OptionValue{IsSome: false}, Type: "Option"}}, nil
}

// option::destroy_some<T>(opt: Option<T>) -> T
func optionDestroySome(_ *l1.CallContext, _ *iotago.ProgrammableMoveCall, args []l1.Value) ([]l1.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("option::destroy_some requires 1 argument")
	}
	opt, ok := args[0].Raw.(*OptionValue)
	if !ok {
		return nil, fmt.Errorf("option::destroy_some: argument is not an Option")
	}
	if !opt.IsSome {
		return nil, fmt.Errorf("option::destroy_some called on None")
	}
	return []l1.Value{*opt.Value}, nil
}

// option::destroy_none<T>(opt: Option<T>)
func optionDestroyNone(_ *l1.CallContext, _ *iotago.ProgrammableMoveCall, _ []l1.Value) ([]l1.Value, error) {
	return nil, nil
}

// option::is_some<T>(opt: &Option<T>) -> bool
func optionIsSome(_ *l1.CallContext, _ *iotago.ProgrammableMoveCall, args []l1.Value) ([]l1.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("option::is_some requires 1 argument")
	}
	opt, ok := args[0].Raw.(*OptionValue)
	if !ok {
		return nil, fmt.Errorf("option::is_some: argument is not an Option")
	}
	return []l1.Value{{Raw: opt.IsSome, Type: "bool"}}, nil
}
