package test

import (
	"testing"

	"github.com/iotaledger/wasp/contracts/wasm/testcore/go/testcore"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/stretchr/testify/require"
)

//nolint:dupl
func TestTypesFull(t *testing.T) {
	run2(t, func(t *testing.T, w bool) {
		ctx := deployTestCore(t, w)

		f := testcore.ScFuncs.PassTypesFull(ctx)
		f.Params.Address().SetValue(ctx.ChainID().Address())
		f.Params.AgentID().SetValue(ctx.Originator().ScAgentID())
		f.Params.ChainID().SetValue(ctx.ChainID())
		f.Params.ContractID().SetValue(ctx.AccountID())
		f.Params.Hash().SetValue(ctx.Cvt.ScHash(hashing.HashStrings("Hash")))
		f.Params.Hname().SetValue(ctx.Cvt.ScHname(iscp.Hn("Hname")))
		f.Params.HnameZero().SetValue(0)
		f.Params.Int64().SetValue(42)
		f.Params.Int64Zero().SetValue(0)
		f.Params.String().SetValue("string")
		f.Params.StringZero().SetValue("")
		f.Func.Post()
		require.NoError(t, ctx.Err)
	})
}

//nolint:dupl
func TestTypesView(t *testing.T) {
	run2(t, func(t *testing.T, w bool) {
		ctx := deployTestCore(t, w)

		v := testcore.ScFuncs.PassTypesView(ctx)
		v.Params.Address().SetValue(ctx.ChainID().Address())
		v.Params.AgentID().SetValue(ctx.Originator().ScAgentID())
		v.Params.ChainID().SetValue(ctx.ChainID())
		v.Params.ContractID().SetValue(ctx.AccountID())
		v.Params.Hash().SetValue(ctx.Cvt.ScHash(hashing.HashStrings("Hash")))
		v.Params.Hname().SetValue(ctx.Cvt.ScHname(iscp.Hn("Hname")))
		v.Params.HnameZero().SetValue(0)
		v.Params.Int64().SetValue(42)
		v.Params.Int64Zero().SetValue(0)
		v.Params.String().SetValue("string")
		v.Params.StringZero().SetValue("")
		v.Func.Call()
		require.NoError(t, ctx.Err)
	})
}
