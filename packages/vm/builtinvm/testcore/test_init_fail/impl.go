package test_init_fail

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/contract"
	"github.com/iotaledger/wasp/packages/vm/examples"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

const (
	Name        = "test_init_fail"
	description = "Dummy contract for testing"
	ParamFail   = "initFailParam"
)

var (
	Interface = &contract.ContractInterface{
		Name:        Name,
		Description: description,
		ProgramHash: *hashing.HashStrings(Name),
	}
)

func init() {
	Interface.WithFunctions(initialize, nil)
	examples.AddProcessor(Interface)
}

func initialize(ctx vmtypes.Sandbox) (dict.Dict, error) {
	if p, err := ctx.Params().Get(ParamFail); err == nil && p != nil {
		return nil, fmt.Errorf("failing on purpose")
	}
	return nil, nil
}
