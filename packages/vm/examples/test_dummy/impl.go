package test_env

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/contract"
	"github.com/iotaledger/wasp/packages/vm/examples"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

const (
	Name        = "helloworld_hc"
	Version     = "0.1"
	fullName    = Name + "-" + Version
	description = "Hardcoded Hello world contract"
)

var (
	Interface = &contract.ContractInterface{
		Name:        fullName,
		Description: description,
		ProgramHash: *hashing.HashStrings(fullName),
	}
)

func init() {
	Interface.WithFunctions(initialize, nil)
}

var (
	ProgramHash, _ = hashing.HashValueFromBase58(description)
)

func init() {
	examples.AddProcessor(ProgramHash, Interface)
}

func initialize(ctx vmtypes.Sandbox) (dict.Dict, error) {
	ctx.Eventf("helloworld-hc.init in %s", ctx.ContractID().Hname().String())
	ctx.Event("Hello World!")

	r, err := ctx.Params().Get("fail")
	if err != nil {
		ctx.Eventf("helloworld-hc.init.fail: %v", err)
		return nil, err
	}
	if r != nil {
		ctx.Eventf("helloworld-hc.init.OK: failing on purpose")
		return nil, fmt.Errorf("helloworld-hc.init.OK: failing on purpose")
	}

	ctx.Eventf("helloworld-hc.init.success")
	return nil, nil
}
