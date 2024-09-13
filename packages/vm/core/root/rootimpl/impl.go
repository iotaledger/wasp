// 'root' a core contract on the chain. It is responsible for:
// - initial setup of the chain during chain deployment
// - maintaining of core parameters of the chain
// - maintaining (setting, delegating) chain owner ID
// - maintaining (granting, revoking) smart contract deployment rights
// - deployment of smart contracts on the chain and maintenance of contract registry

package rootimpl

import (
	"github.com/samber/lo"

	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/errors/coreerrors"
	"github.com/iotaledger/wasp/packages/vm/core/root"
)

var Processor = root.Contract.Processor(nil,
	root.FuncGrantDeployPermission.WithHandler(grantDeployPermission),
	root.FuncRequireDeployPermissions.WithHandler(requireDeployPermissions),
	root.FuncRevokeDeployPermission.WithHandler(revokeDeployPermission),
	root.ViewFindContract.WithHandler(findContract),
	root.ViewGetContractRecords.WithHandler(getContractRecords),
)

var errInvalidContractName = coreerrors.Register("invalid contract name").Create()

// deployContract deploys contract and calls its 'init' constructor.
// If call to the constructor returns an error or another error occurs,
// removes smart contract form the registry as if it was never attempted to deploy
// Inputs:
//   - ParamName string, the unique name of the contract in the chain. Later used as Hname
//   - ParamProgramHash HashValue is a hash of the blob which represents program binary in the 'blob' contract.
//     In case of hardcoded examples it's an arbitrary unique hash set in the global call examples.AddProcessor
func deployContract(ctx isc.Sandbox, progHash hashing.HashValue, name string, initArgs isc.CallArguments) {
	ctx.Log().Debugf("root.deployContract.begin")
	if !isAuthorizedToDeploy(ctx) {
		panic(vm.ErrUnauthorized)
	}

	if name == "" || len(name) > 255 {
		panic(errInvalidContractName)
	}

	// call to load VM from binary to check if it loads successfully
	err := ctx.Privileged().TryLoadContract(progHash)
	ctx.RequireNoError(err, "root.deployContract.fail 1: ")

	// VM loaded successfully. Storing contract in the registry and calling constructor
	state := root.NewStateWriterFromSandbox(ctx)
	state.StoreContractRecord(&root.ContractRecord{
		ProgramHash: progHash,
		Name:        name,
	})
	ctx.Call(isc.NewMessage(isc.Hn(name), isc.EntryPointInit, initArgs), nil)
	eventDeploy(ctx, progHash, name)
}

// grantDeployPermission grants permission to deploy contracts
func grantDeployPermission(ctx isc.Sandbox, deployer isc.AgentID) {
	ctx.RequireCallerIsChainOwner()
	state := root.NewStateWriterFromSandbox(ctx)
	state.GetDeployPermissions().SetAt(deployer.Bytes(), []byte{0x01})
	eventGrant(ctx, deployer)
}

// revokeDeployPermission revokes permission to deploy contracts
func revokeDeployPermission(ctx isc.Sandbox, deployer isc.AgentID) {
	ctx.RequireCallerIsChainOwner()
	state := root.NewStateWriterFromSandbox(ctx)
	state.GetDeployPermissions().DelAt(deployer.Bytes())
	eventRevoke(ctx, deployer)
}

func requireDeployPermissions(ctx isc.Sandbox, permissionsEnabled bool) {
	ctx.RequireCallerIsChainOwner()
	state := root.NewStateWriterFromSandbox(ctx)
	state.SetDeployPermissionsEnabled(permissionsEnabled)
}

// findContract view finds and returns encoded record of the contract
func findContract(ctx isc.SandboxView, hname isc.Hname) (bool, **root.ContractRecord) {
	state := root.NewStateReaderFromSandbox(ctx)
	rec := state.FindContract(hname)
	if rec == nil {
		return false, nil
	}
	return true, &rec
}

func getContractRecords(ctx isc.SandboxView) []lo.Tuple2[*isc.Hname, *root.ContractRecord] {
	var ret []lo.Tuple2[*isc.Hname, *root.ContractRecord]
	state := root.NewStateReaderFromSandbox(ctx)
	state.GetContractRegistry().Iterate(func(elemKey []byte, value []byte) bool {
		hname := codec.Hname.MustDecode(elemKey)
		rec := lo.Must(root.ContractRecordFromBytes(value))
		ret = append(ret, lo.T2(&hname, rec))
		return true
	})
	return ret
}
