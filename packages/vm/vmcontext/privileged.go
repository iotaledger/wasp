package vmcontext

import (
	"fmt"
	"math/big"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"golang.org/x/xerrors"

	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/vm/core/root"
)

func (vmctx *VMContext) mustBeCalledFromContract(contract *coreutil.ContractInfo) {
	if vmctx.CurrentContractHname() != contract.Hname() {
		panic(fmt.Sprintf("%v: core contract '%s' expected", ErrPriviledgedCallFailed, contract.Name))
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

func (vmctx *VMContext) DestroyFoundry(sn uint32) int64 {
	vmctx.mustBeCalledFromContract(accounts.Contract)
	panic("DestroyFoundry: implement me")
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
