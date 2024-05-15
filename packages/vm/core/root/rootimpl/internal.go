package rootimpl

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/core/root"
)

// isAuthorizedToDeploy checks if caller is authorized to deploy smart contract
func isAuthorizedToDeploy(ctx isc.Sandbox) bool {
	state := root.NewStateReaderFromSandbox(ctx)
	permissionsEnabled := state.GetDeployPermissionsEnabled()
	if !permissionsEnabled {
		return true
	}

	caller := ctx.Caller()
	if caller.Equals(ctx.ChainOwnerID()) {
		// chain owner is always authorized
		return true
	}
	if ctx.ChainID().IsSameChain(caller) {
		// smart contract from the same chain is always authorized
		return true
	}

	return state.GetDeployPermissions().HasAt(caller.Bytes())
}
