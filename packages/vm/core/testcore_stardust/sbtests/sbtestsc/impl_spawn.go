package sbtestsc

import (
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

// spawn deploys new contract and calls it
func spawn(ctx iscp.Sandbox) dict.Dict {
	ctx.Log().Debugf(FuncSpawn.Name)
	name := Contract.Name + "_spawned"
	dscr := "spawned contract description"
	hname := iscp.Hn(name)
	ctx.DeployContract(Contract.ProgramHash, name, dscr, nil)

	for i := 0; i < 5; i++ {
		ctx.Call(hname, FuncIncCounter.Hname(), nil, nil)
	}
	ctx.Log().Debugf("sbtestsc.spawn: new contract name = %s hname = %s", name, hname.String())
	return nil
}
