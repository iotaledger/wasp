package vmcontext

import (
	"math/big"

	iotago "github.com/iotaledger/iota.go/v3"
)

func (vmctx *VMContext) CreateNewFoundry(scheme iotago.TokenScheme, tag iotago.TokenTag, maxSupply *big.Int) iotago.NativeTokenID {
	panic("not implemented")
	//serNum := vmctx.txbuilder.CreateNewFoundry(scheme, tag, maxSupply)

	// TODO
	//
	//vmctx.pushCallContext(accounts.Contract.Hname(), nil, nil)
	//accounts.AddFoundry(vmctx.State(), vmctx.Caller(), serNum)
	//vmctx.popCallContext()
	//
	//vmctx.pushCallContext(blocklog.Contract.Hname(), nil, nil)
	//blocklog.GetFoundryOutput(vmctx.State(), serNum)
	//vmctx.popCallContext()

}

func (vmctx *VMContext) DestroyFoundry(id iotago.NativeTokenID) {
	panic("implement me")
}

func (vmctx *VMContext) Supply(id iotago.NativeTokenID) (*big.Int, *big.Int) {
	panic("implement me")
}

func (vmctx *VMContext) ModifySupply(id iotago.NativeTokenID, delta *big.Int) {
	panic("implement me")
}
