package vmcontext

import (
	"fmt"
	"math/big"

	"github.com/iotaledger/wasp/packages/vm"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/execution"
	"golang.org/x/xerrors"
)

func (vmctx *VMContext) mustBeCalledFromContract(contract *coreutil.ContractInfo) {
	if vmctx.CurrentContractHname() != contract.Hname() {
		panic(fmt.Sprintf("%v: core contract '%s' expected", vm.ErrPrivilegedCallFailed, contract.Name))
	}
}

func (vmctx *VMContext) TryLoadContract(programHash hashing.HashValue) error {
	vmctx.mustBeCalledFromContract(root.Contract)
	vmtype, programBinary, err := execution.GetProgramBinary(vmctx, programHash)
	if err != nil {
		return err
	}
	return vmctx.task.Processors.NewProcessor(programHash, programBinary, vmtype)
}

func (vmctx *VMContext) CreateNewFoundry(scheme iotago.TokenScheme, tag iotago.TokenTag, metadata []byte) (uint32, uint64) {
	vmctx.mustBeCalledFromContract(accounts.Contract)
	return vmctx.txbuilder.CreateNewFoundry(scheme, tag, metadata)
}

func (vmctx *VMContext) DestroyFoundry(sn uint32) uint64 {
	vmctx.mustBeCalledFromContract(accounts.Contract)
	return vmctx.txbuilder.DestroyFoundry(sn)
}

func (vmctx *VMContext) ModifyFoundrySupply(sn uint32, delta *big.Int) int64 {
	vmctx.mustBeCalledFromContract(accounts.Contract)
	out, _, _ := accounts.GetFoundryOutput(vmctx.State(), sn, vmctx.ChainID())
	tokenID, err := out.NativeTokenID()
	if err != nil {
		panic(xerrors.Errorf("internal: %v", err))
	}
	return vmctx.txbuilder.ModifyNativeTokenSupply(&tokenID, delta)
}

func (vmctx *VMContext) BlockContext(ctx iscp.Sandbox, construct func(ctx iscp.Sandbox) interface{}, onClose func(interface{})) interface{} {
	hname := vmctx.CurrentContractHname()
	if bctx, alreadyExists := vmctx.blockContext[hname]; alreadyExists {
		return bctx.obj
	}
	if onClose == nil {
		onClose = func(interface{}) {}
	}
	ret := &blockContext{
		obj:     construct(ctx),
		onClose: onClose,
	}
	vmctx.blockContext[hname] = ret
	// storing sequence to have deterministic order of closing
	vmctx.blockContextCloseSeq = append(vmctx.blockContextCloseSeq, hname)
	return ret.obj
}
