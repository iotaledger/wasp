package vmcontext

import (
	"math/big"

	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"golang.org/x/xerrors"

	iotago "github.com/iotaledger/iota.go/v3"
)

func (vmctx *VMContext) CreateNewFoundry(scheme iotago.TokenScheme, tag iotago.TokenTag, maxSupply *big.Int, metadata []byte) (uint32, uint64) {
	return vmctx.txbuilder.CreateNewFoundry(scheme, tag, maxSupply, metadata)
}

func (vmctx *VMContext) DestroyFoundry(sn uint32) int64 {
	panic("DestroyFoundry: implement me")
}

func (vmctx *VMContext) GetOutput(sn uint32) *iotago.FoundryOutput {
	ret, _, _ := accounts.GetFoundryOutput(vmctx.State(), sn, vmctx.ChainID())
	return ret
}

func (vmctx *VMContext) ModifySupply(sn uint32, delta *big.Int) int64 {
	out := vmctx.GetOutput(sn)
	tokenID, err := out.NativeTokenID()
	if err != nil {
		panic(xerrors.Errorf("internal: %v", err))
	}
	return vmctx.txbuilder.ModifyNativeTokenSupply(&tokenID, delta)
}
