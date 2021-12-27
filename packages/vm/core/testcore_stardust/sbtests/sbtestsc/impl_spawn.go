package sbtestsc

import (
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

// spawn deploys new contract and calls it
func spawn(ctx iscp.Sandbox) (dict.Dict, error) {
	ctx.Log().Debugf(FuncSpawn.Name)
	name := Contract.Name + "_spawned"
	dscr := "spawned contract description"
	hname := iscp.Hn(name)
	err := ctx.DeployContract(Contract.ProgramHash, name, dscr, nil)
	if err != nil {
		return nil, err
	}

	for i := 0; i < 5; i++ {
		_, err := ctx.Call(hname, FuncIncCounter.Hname(), nil, nil)
		if err != nil {
			return nil, err
		}
	}
	ctx.Log().Debugf("sbtestsc.spawn: new contract name = %s hname = %s", name, hname.String())
	return nil, nil
}
