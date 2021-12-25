package vmcontext

import (
	"math/big"

	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"golang.org/x/xerrors"

	iotago "github.com/iotaledger/iota.go/v3"
)

func (vmctx *VMContext) CreateNewFoundry(scheme iotago.TokenScheme, tag iotago.TokenTag, maxSupply *big.Int) uint32 {
	return vmctx.txbuilder.CreateNewFoundry(scheme, tag, maxSupply)
}

func (vmctx *VMContext) DestroyFoundry(sn uint32) {
	panic("implement me")
}

func (vmctx *VMContext) GetOutput(sn uint32) *iotago.FoundryOutput {
	ret, _, _ := accounts.GetFoundryOutput(vmctx.State(), sn, vmctx.ChainID())
	return ret
}

func (vmctx *VMContext) ModifySupply(sn uint32, delta *big.Int) {
	out := vmctx.GetOutput(sn)
	tokenID, err := out.NativeTokenID()
	if err != nil {
		panic(xerrors.Errorf("internal: %w", err))
	}
	vmctx.txbuilder.ModifyNativeTokenSupply(&tokenID, delta)
}
