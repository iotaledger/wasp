package rootimpl

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/vm/core/root"
)

// isAuthorizedToDeploy checks if caller is authorized to deploy smart contract
func isAuthorizedToDeploy(ctx isc.Sandbox) bool {
	permissionsEnabled, err := codec.DecodeBool(ctx.State().MustGet(root.StateVarDeployPermissionsEnabled))
	if err != nil {
		return false
	}
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

	return collections.NewMap(ctx.State(), root.StateVarDeployPermissions).MustHasAt(caller.Bytes())
}

func storeContractRecord(state kv.KVStore, rec *root.ContractRecord) {
	hname := isc.Hn(rec.Name)
	// storing contract record in the registry
	contractRegistry := root.GetContractRegistry(state)
	if contractRegistry.MustHasAt(hname.Bytes()) {
		panic(fmt.Sprintf("contract '%s'/%s already exists", rec.Name, hname.String()))
	}
	contractRegistry.MustSetAt(hname.Bytes(), rec.Bytes())
}
