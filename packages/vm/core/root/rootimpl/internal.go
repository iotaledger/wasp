package rootimpl

import (
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/root"
)

func mustStoreContract(ctx iscp.Sandbox, i *coreutil.ContractInfo) {
	rec := root.NewContractRecord(i, &iscp.NilAgentID)
	ctx.Log().Debugf("mustStoreAndInitCoreContract: '%s', hname = %s", i.Name, i.Hname())
	mustStoreContractRecord(ctx, rec)
}

func mustStoreAndInitCoreContract(ctx iscp.Sandbox, i *coreutil.ContractInfo, params ...dict.Dict) {
	mustStoreContract(ctx, i)
	var p dict.Dict
	if len(params) == 1 {
		p = params[0]
	}
	_, err := ctx.Call(iscp.Hn(i.Name), iscp.EntryPointInit, p, nil)
	ctx.RequireNoError(err)
}

func mustStoreContractRecord(ctx iscp.Sandbox, rec *root.ContractRecord) {
	hname := rec.Hname()
	contractRegistry := collections.NewMap(ctx.State(), root.StateVarContractRegistry)
	ctx.Require(!contractRegistry.MustHasAt(hname.Bytes()), "contract '%s'/%s already exist", rec.Name, hname.String())
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

func isChainOwner(ctx iscp.Sandbox) bool {
	ret, err := ctx.Call(governance.Contract.Hname(), governance.FuncGetChainOwner.Hname(), nil, nil)
	ctx.RequireNoError(err)
	owner, err := codec.DecodeAgentID(ret.MustGet(governance.ParamChainOwner))
	ctx.RequireNoError(err)
	return owner.Equals(ctx.Caller())
}
