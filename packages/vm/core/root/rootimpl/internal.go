package rootimpl

import (
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/root"
)

func mustStoreContract(ctx iscp.Sandbox, i *coreutil.ContractInfo) {
	rec := root.NewContractRecord(i, &iscp.NilAgentID)
	ctx.Log().Debugf("mustStoreAndInitCoreContract: '%s', hname = %s", i.Name, i.Hname())
	mustStoreContractRecord(ctx, rec)
}

func mustStoreAndInitCoreContract(ctx iscp.Sandbox, i *coreutil.ContractInfo, params dict.Dict) {
	mustStoreContract(ctx, i)
	ctx.Call(iscp.Hn(i.Name), iscp.EntryPointInit, params, nil)
}

func mustStoreContractRecord(ctx iscp.Sandbox, rec *root.ContractRecord) {
	hname := rec.Hname()
	contractRegistry := root.GetContractRegistry(ctx.State())
	ctx.Requiref(!contractRegistry.MustHasAt(hname.Bytes()), "contract '%s'/%s already exist", rec.Name, hname.String())
	contractRegistry.MustSetAt(hname.Bytes(), rec.Bytes())
}

// isAuthorizedToDeploy checks if caller is authorized to deploy smart contract
func isAuthorizedToDeploy(ctx iscp.Sandbox) bool {
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
	if caller.Address().Equal(ctx.ChainID().AsAddress()) {
		// smart contract from the same chain is always authorized
		return true
	}

	return collections.NewMap(ctx.State(), root.StateVarDeployPermissions).MustHasAt(caller.Bytes())
}
