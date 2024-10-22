package sbtestsc

import (
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
)

// spawn deploys new contract and calls it
func spawn(ctx isc.Sandbox, progHash hashing.HashValue) {
	ctx.Log().Debugf(FuncSpawn.Name)
	name := Contract.Name + "_spawned"
	hname := isc.Hn(name)

	panic("TODO: contract deployment")
	//ctx.DeployContract(progHash, name, nil)

	for i := 0; i < 5; i++ {
		ctx.Call(isc.NewMessage(hname, FuncIncCounter.Hname(), nil), nil)
	}
	ctx.Log().Debugf("sbtestsc.spawn: new contract name = %s hname = %s", name, hname.String())
}
