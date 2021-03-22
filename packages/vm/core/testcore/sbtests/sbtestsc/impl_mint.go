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
	h := ctx.Utils().Hashing().Blake2b(ctx.RequestID().Bytes())
	color, _, err := ledgerstate.ColorFromBytes(h[:])
	a.RequireNoError(err)
	supply, _ := allMinted[color]
	ret.Set(VarMintedSupply, codec.EncodeUint64(supply))
	return ret, nil
}
