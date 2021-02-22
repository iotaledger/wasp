package sbtestsc

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

func getMintedSupply(ctx coretypes.Sandbox) (dict.Dict, error) {
	ret := dict.New()
	ret.Set(VarMintedSupply, codec.EncodeInt64(ctx.MintedSupply()))
	return ret, nil
}
