package rootimpl

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/root"
)

func storeCoreContract(ctx isc.Sandbox, i *coreutil.ContractInfo) {
	rec := root.ContractRecordFromContractInfo(i)
	ctx.Log().Debugf("storeCoreContract: '%s', hname = %s", i.Name, i.Hname())
	storeContractRecord(ctx, rec)
}

func storeAndInitCoreContract(ctx isc.Sandbox, i *coreutil.ContractInfo, params dict.Dict) {
	storeCoreContract(ctx, i)
	ctx.Call(isc.Hn(i.Name), isc.EntryPointInit, params, nil)
}

func storeContractRecord(ctx isc.Sandbox, rec *root.ContractRecord) {
	hname := isc.Hn(rec.Name)
	// storing contract record in the registry
	contractRegistry := root.GetContractRegistry(ctx.State())
	ctx.Requiref(!contractRegistry.MustHasAt(hname.Bytes()), "contract '%s'/%s already exists", rec.Name, hname.String())
	contractRegistry.MustSetAt(hname.Bytes(), rec.Bytes())
}

// isAuthorizedToDeploy checks if caller is authorized to deploy smart contract
func isAuthorizedToDeploy(ctx isc.Sandbox) bool {
	permissionsEnabled, err := codec.DecodeBool(ctx.State().MustGet(root.StateVarDeployPermissionsEnabled), false)
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
