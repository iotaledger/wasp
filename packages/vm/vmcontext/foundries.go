package vmcontext

import (
	"math/big"

	iotago "github.com/iotaledger/iota.go/v3"
)

func (vmctx *VMContext) CreateNewFoundry(scheme iotago.TokenScheme, tag iotago.TokenTag, maxSupply *big.Int) uint32 {
	return vmctx.txbuilder.CreateNewFoundry(scheme, tag, maxSupply)
}

func (vmctx *VMContext) DestroyFoundry(sn uint32) {
	panic("implement me")
}

func (vmctx *VMContext) GetOutput(sn uint32) *iotago.FoundryOutput {
	panic("implement me")
}

func (vmctx *VMContext) ModifySupply(sn uint32, delta *big.Int) {
	panic("implement me")
}
