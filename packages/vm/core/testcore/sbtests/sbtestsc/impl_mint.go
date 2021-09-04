package sbtestsc

import (
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/assert"
	"github.com/iotaledger/wasp/packages/iscp/colored"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

func getMintedSupply(ctx iscp.Sandbox) (dict.Dict, error) {
	ret := dict.New()
	allMinted := ctx.Minted()
	a := assert.NewAssert(ctx.Log())
	a.Require(len(allMinted) == 1, "test only supports one minted color")
	var colMinted colored.Color
	var amount uint64
	var err error
	for col, bal := range allMinted {
		colMinted, err = colored.NewColor(col)
		a.RequireNoError(err, "creating color from key")
		amount = bal
	}
	ret.Set(VarMintedSupply, codec.EncodeUint64(amount))
	ret.Set(VarMintedColor, codec.EncodeColor(colMinted))
	return ret, nil
}
