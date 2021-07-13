package sbtestsc

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

// spawn deploys new contract and calls it
func spawn(ctx coretypes.Sandbox) (dict.Dict, error) {
	ctx.Log().Debugf(FuncSpawn)
	name := Interface.Name + "_spawned"
	dscr := "spawned contract description"
	hname := coretypes.Hn(name)
	err := ctx.DeployContract(Interface.ProgramHash, name, dscr, nil)
	if err != nil {
		return nil, err
	}

	for i := 0; i < 5; i++ {
		_, err := ctx.Call(hname, coretypes.Hn(FuncIncCounter), nil, nil)
		if err != nil {
			return nil, err
		}
	}
	ctx.Log().Debugf("sbtestsc.spawn: new contract name = %s hname = %s", name, hname.String())
	return nil, nil
}
