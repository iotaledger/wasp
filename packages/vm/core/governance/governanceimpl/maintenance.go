package governanceimpl

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
)

// Maintenance mode means no requests will be processed except calls to the governance contract
// NOTE: Maintenance mode is not available if the governing address is a Contract on the chain itself. (otherwise setting maintenance ON will result in a deadlock)

func startMaintenance(ctx isc.Sandbox) dict.Dict {
	ctx.RequireCallerIsChainOwner()
	// check if caller is a contract from this chain, panic if so.
	caller := ctx.Caller()
	if caller.Kind() == isc.AgentIDKindContract &&
		caller.(*isc.ContractAgentID).ChainID().Equals(ctx.ChainID()) {
		panic(vm.ErrUnauthorized)
	}
	state := governance.NewStateWriterFromSandbox(ctx)
	state.SetMaintenanceStatus(true)
	return nil
}

func stopMaintenance(ctx isc.Sandbox) dict.Dict {
	ctx.RequireCallerIsChainOwner()
	state := governance.NewStateWriterFromSandbox(ctx)
	state.SetMaintenanceStatus(false)
	return nil
}

func getMaintenanceStatus(ctx isc.SandboxView) bool {
	state := governance.NewStateReaderFromSandbox(ctx)
	return state.GetMaintenanceStatus()
}
