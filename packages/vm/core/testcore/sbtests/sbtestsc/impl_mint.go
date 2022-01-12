package sbtestsc

import (
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

func getMintedSupply(ctx iscp.Sandbox) (dict.Dict, error) {
	panic("not implemented")
	// ret := dict.New()
	// allMinted := ctx.Minted()
	// a := assert.NewAssert(ctx.Log())
	// a.Requiref(len(allMinted) == 1, "test supports only one minted color")
	// var colMinted colored.Color
	// var amount uint64
	// for col, bal := range allMinted {
	// 	colMinted = col
	// 	amount = bal
	// }
	// ret.Set(VarMintedSupply, codec.EncodeUint64(amount))
	// ret.Set(VarMintedColor, codec.EncodeColor(colMinted))
	// return ret, nil
}
