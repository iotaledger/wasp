package governanceimpl

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
)

// Maintenance mode means no requests will be processed except calls to the governance contract
// NOTE: Maintenance mode is not available if the governing address is a Contract on the chain itself. (otherwise setting maintence ON will result in a deadlock)

func setMaintenanceOn(ctx isc.Sandbox) dict.Dict {
	ctx.RequireCallerIsChainOwner()
	// check if caller is a contract from this chain, panic if so.
	if ctx.Caller().Kind() == isc.AgentIDKindContract &&
		ctx.Caller().(*isc.ContractAgentID).ChainID().Equals(ctx.ChainID()) {
		panic(vm.ErrUnauthorized)
	}
	ctx.State().Set(governance.VarMaintenanceStatus, codec.Encode(true))
	return nil
}

func setMaintenanceOff(ctx isc.Sandbox) dict.Dict {
	ctx.RequireCallerIsChainOwner()
	ctx.State().Set(governance.VarMaintenanceStatus, codec.Encode(false))
	return nil
}

func getMaintenanceStatus(ctx isc.SandboxView) dict.Dict {
	return dict.Dict{
		governance.VarMaintenanceStatus: ctx.State().MustGet(governance.VarMaintenanceStatus),
	}
}
