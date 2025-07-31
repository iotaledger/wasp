package governanceimpl

import (
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/vm"
	"github.com/iotaledger/wasp/v2/packages/vm/core/governance"
)

// Maintenance mode means no requests will be processed except calls to the governance contract
// NOTE: Maintenance mode is not available if the governing address is a Contract on the chain itself. (otherwise setting maintenance ON will result in a deadlock)

func startMaintenance(ctx isc.Sandbox) {
	ctx.RequireCallerIsChainAdmin()
	// check if caller is a contract from this chain, panic if so.
	caller := ctx.Caller()
	if caller.Kind() == isc.AgentIDKindContract {
		panic(vm.ErrUnauthorized)
	}
	state := governance.NewStateWriterFromSandbox(ctx)
	state.SetMaintenanceStatus(true)
}

func stopMaintenance(ctx isc.Sandbox) {
	ctx.RequireCallerIsChainAdmin()
	state := governance.NewStateWriterFromSandbox(ctx)
	state.SetMaintenanceStatus(false)
}

func getMaintenanceStatus(ctx isc.SandboxView) bool {
	state := governance.NewStateReaderFromSandbox(ctx)
	return state.GetMaintenanceStatus()
}
