package sbtestsc

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/assert"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

func getMintedSupply(ctx coretypes.Sandbox) (dict.Dict, error) {
	ret := dict.New()
	allMinted := ctx.Minted()
	a := assert.NewAssert(ctx.Log())
	a.Require(len(allMinted) == 1, "test only supports one minted color")
	var color ledgerstate.Color
	var amount uint64
	for col, bal := range allMinted {
		color = col
		amount = bal
	}
	ret.Set(VarMintedSupply, codec.EncodeUint64(amount))
	ret.Set(VarMintedColor, codec.EncodeColor(color))
	return ret, nil
}
