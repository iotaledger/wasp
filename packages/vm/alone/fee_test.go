package alone

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/accountsc"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/blob"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestFeeBasic(t *testing.T) {
	glb := New(t, false, false)
	chain := glb.NewChain(nil, "chain1")

	col, fee := chain.GetFeeInfo(root.Interface.Hname())
	require.EqualValues(t, balance.ColorIOTA, col)
	require.EqualValues(t, 0, fee)

	col, fee = chain.GetFeeInfo(accountsc.Interface.Hname())
	require.EqualValues(t, balance.ColorIOTA, col)
	require.EqualValues(t, 0, fee)

	col, fee = chain.GetFeeInfo(blob.Interface.Hname())
	require.EqualValues(t, balance.ColorIOTA, col)
	require.EqualValues(t, 0, fee)
}

func TestSetDefaultFeeOk(t *testing.T) {
	glb := New(t, false, false)
	chain := glb.NewChain(nil, "chain1")

	req := NewCall(root.Interface.Name, root.FuncSetDefaultFee, root.ParamDefaultFee, 1000)
	_, err := chain.PostRequest(req, nil)
	require.NoError(t, err)

	col, fee := chain.GetFeeInfo(root.Interface.Hname())
	require.EqualValues(t, balance.ColorIOTA, col)
	require.EqualValues(t, 1000, fee)

	col, fee = chain.GetFeeInfo(accountsc.Interface.Hname())
	require.EqualValues(t, balance.ColorIOTA, col)
	require.EqualValues(t, 1000, fee)

	col, fee = chain.GetFeeInfo(blob.Interface.Hname())
	require.EqualValues(t, balance.ColorIOTA, col)
	require.EqualValues(t, 1000, fee)
}

func TestSetDefaultFeeFailNegative(t *testing.T) {
	glb := New(t, false, false)
	chain := glb.NewChain(nil, "chain1")

	req := NewCall(root.Interface.Name, root.FuncSetDefaultFee, root.ParamDefaultFee, -2)
	_, err := chain.PostRequest(req, nil)
	require.Error(t, err)

	col, fee := chain.GetFeeInfo(root.Interface.Hname())
	require.EqualValues(t, balance.ColorIOTA, col)
	require.EqualValues(t, 0, fee)

	col, fee = chain.GetFeeInfo(accountsc.Interface.Hname())
	require.EqualValues(t, balance.ColorIOTA, col)
	require.EqualValues(t, 0, fee)

	col, fee = chain.GetFeeInfo(blob.Interface.Hname())
	require.EqualValues(t, balance.ColorIOTA, col)
	require.EqualValues(t, 0, fee)
}

func TestSetDefaultFeeNotAuthorized(t *testing.T) {
	glb := New(t, false, false)
	chain := glb.NewChain(nil, "chain1")

	user := glb.NewSigSchemeWithFunds()

	req := NewCall(root.Interface.Name, root.FuncSetDefaultFee, root.ParamDefaultFee, 1000)
	_, err := chain.PostRequest(req, user)
	require.Error(t, err)

	col, fee := chain.GetFeeInfo(root.Interface.Hname())
	require.EqualValues(t, balance.ColorIOTA, col)
	require.EqualValues(t, 0, fee)

	col, fee = chain.GetFeeInfo(accountsc.Interface.Hname())
	require.EqualValues(t, balance.ColorIOTA, col)
	require.EqualValues(t, 0, fee)

	col, fee = chain.GetFeeInfo(blob.Interface.Hname())
	require.EqualValues(t, balance.ColorIOTA, col)
	require.EqualValues(t, 0, fee)
}

func TestSetFeeOk(t *testing.T) {
	glb := New(t, false, false)
	chain := glb.NewChain(nil, "chain1")

	req := NewCall(root.Interface.Name, root.FuncSetFee,
		root.ParamHname, blob.Interface.Hname(),
		root.ParamContractFee, 1000,
	)
	_, err := chain.PostRequest(req, nil)
	require.NoError(t, err)

	col, fee := chain.GetFeeInfo(root.Interface.Hname())
	require.EqualValues(t, balance.ColorIOTA, col)
	require.EqualValues(t, 0, fee)

	col, fee = chain.GetFeeInfo(accountsc.Interface.Hname())
	require.EqualValues(t, balance.ColorIOTA, col)
	require.EqualValues(t, 0, fee)

	col, fee = chain.GetFeeInfo(blob.Interface.Hname())
	require.EqualValues(t, balance.ColorIOTA, col)
	require.EqualValues(t, 1000, fee)
}

func TestSetFeeNotAuthorized(t *testing.T) {
	glb := New(t, false, false)
	chain := glb.NewChain(nil, "chain1")

	user := glb.NewSigSchemeWithFunds()
	req := NewCall(root.Interface.Name, root.FuncSetFee,
		root.ParamHname, blob.Interface.Hname(),
		root.ParamContractFee, 1000,
	)
	_, err := chain.PostRequest(req, user)
	require.Error(t, err)

	col, fee := chain.GetFeeInfo(root.Interface.Hname())
	require.EqualValues(t, balance.ColorIOTA, col)
	require.EqualValues(t, 0, fee)

	col, fee = chain.GetFeeInfo(accountsc.Interface.Hname())
	require.EqualValues(t, balance.ColorIOTA, col)
	require.EqualValues(t, 0, fee)

	col, fee = chain.GetFeeInfo(blob.Interface.Hname())
	require.EqualValues(t, balance.ColorIOTA, col)
	require.EqualValues(t, 0, fee)
}

func TestSetFeeWithDefault(t *testing.T) {
	glb := New(t, false, false)
	chain := glb.NewChain(nil, "chain1")

	req := NewCall(root.Interface.Name, root.FuncSetFee,
		root.ParamHname, blob.Interface.Hname(),
		root.ParamContractFee, 1000,
	)
	_, err := chain.PostRequest(req, nil)
	require.NoError(t, err)

	col, fee := chain.GetFeeInfo(root.Interface.Hname())
	require.EqualValues(t, balance.ColorIOTA, col)
	require.EqualValues(t, 0, fee)

	col, fee = chain.GetFeeInfo(accountsc.Interface.Hname())
	require.EqualValues(t, balance.ColorIOTA, col)
	require.EqualValues(t, 0, fee)

	col, fee = chain.GetFeeInfo(blob.Interface.Hname())
	require.EqualValues(t, balance.ColorIOTA, col)
	require.EqualValues(t, 1000, fee)

	req = NewCall(root.Interface.Name, root.FuncSetDefaultFee, root.ParamDefaultFee, 499)
	_, err = chain.PostRequest(req, nil)
	require.NoError(t, err)

	col, fee = chain.GetFeeInfo(root.Interface.Hname())
	require.EqualValues(t, balance.ColorIOTA, col)
	require.EqualValues(t, 499, fee)

	col, fee = chain.GetFeeInfo(accountsc.Interface.Hname())
	require.EqualValues(t, balance.ColorIOTA, col)
	require.EqualValues(t, 499, fee)

	col, fee = chain.GetFeeInfo(blob.Interface.Hname())
	require.EqualValues(t, balance.ColorIOTA, col)
	require.EqualValues(t, 1000, fee)
}
