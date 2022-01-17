package vmcontext

import (
	"fmt"
	"math/big"

	"github.com/iotaledger/wasp/packages/iscp"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"golang.org/x/xerrors"

	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/vm/core/root"
)

func (vmctx *VMContext) mustBeCalledFromContract(contract *coreutil.ContractInfo) {
	if vmctx.CurrentContractHname() != contract.Hname() {
		panic(fmt.Sprintf("%v: core contract '%s' expected", ErrPrivilegedCallFailed, contract.Name))
	}
}

func (vmctx *VMContext) TryLoadContract(programHash hashing.HashValue) error {
	vmctx.mustBeCalledFromContract(root.Contract)
	vmtype, programBinary, err := vmctx.getBinary(programHash)
	if err != nil {
		return err
	}
	return vmctx.task.Processors.NewProcessor(programHash, programBinary, vmtype)
}

func (vmctx *VMContext) CreateNewFoundry(scheme iotago.TokenScheme, tag iotago.TokenTag, maxSupply *big.Int, metadata []byte) (uint32, uint64) {
	vmctx.mustBeCalledFromContract(accounts.Contract)
	return vmctx.txbuilder.CreateNewFoundry(scheme, tag, maxSupply, metadata)
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

// TODO move BlockContext to Privileged sub-interface.
//   do we need ctx as parameter?
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
