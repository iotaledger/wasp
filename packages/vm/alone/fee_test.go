package alone

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/accountsc"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/blob"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
	"github.com/stretchr/testify/require"
	"testing"
)

func checkFees(chain *Chain, contract string, expectedOf, expectedVf int64) {
	col, ownerFee, validatorFee := chain.GetFeeInfo(coretypes.Hn(contract))
	require.EqualValues(chain.Glb.T, balance.ColorIOTA, col)
	require.EqualValues(chain.Glb.T, expectedOf, ownerFee)
	require.EqualValues(chain.Glb.T, expectedVf, validatorFee)
}

func TestFeeBasic(t *testing.T) {
	glb := New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	checkFees(chain, root.Interface.Name, 0, 0)
	checkFees(chain, accountsc.Interface.Name, 0, 0)
	checkFees(chain, blob.Interface.Name, 0, 0)
}

func TestSetDefaultFeeNotAuthorized(t *testing.T) {
	glb := New(t, false, false)
	chain := glb.NewChain(nil, "chain1")

	user := glb.NewSigSchemeWithFunds()

	req := NewCall(root.Interface.Name, root.FuncSetDefaultFee, root.ParamOwnerFee, 1000)
	_, err := chain.PostRequest(req, user)
	require.Error(t, err)

	checkFees(chain, root.Interface.Name, 0, 0)
	checkFees(chain, accountsc.Interface.Name, 0, 0)
	checkFees(chain, blob.Interface.Name, 0, 0)
}

func TestSetContractFeeNotAuthorized(t *testing.T) {
	glb := New(t, false, false)
	chain := glb.NewChain(nil, "chain1")

	user := glb.NewSigSchemeWithFunds()

	req := NewCall(root.Interface.Name, root.FuncSetContractFee, root.ParamOwnerFee, 1000)
	_, err := chain.PostRequest(req, user)
	require.Error(t, err)

	checkFees(chain, root.Interface.Name, 0, 0)
	checkFees(chain, accountsc.Interface.Name, 0, 0)
	checkFees(chain, blob.Interface.Name, 0, 0)
}

func TestSetDefaultOwnerFeeOk(t *testing.T) {
	glb := New(t, false, false)
	chain := glb.NewChain(nil, "chain1")

	req := NewCall(root.Interface.Name, root.FuncSetDefaultFee,
		root.ParamOwnerFee, 1000,
	)
	_, err := chain.PostRequest(req, nil)
	require.NoError(t, err)
	checkFees(chain, root.Interface.Name, 1000, 0)
	checkFees(chain, accountsc.Interface.Name, 1000, 0)
	checkFees(chain, blob.Interface.Name, 1000, 0)
}

func TestSetDefaultValidatorFeeOk(t *testing.T) {
	glb := New(t, false, false)
	chain := glb.NewChain(nil, "chain1")

	req := NewCall(root.Interface.Name, root.FuncSetDefaultFee,
		root.ParamValidatorFee, 499,
	)
	_, err := chain.PostRequest(req, nil)
	require.NoError(t, err)
	checkFees(chain, root.Interface.Name, 0, 499)
	checkFees(chain, accountsc.Interface.Name, 0, 499)
	checkFees(chain, blob.Interface.Name, 0, 499)
}

func TestSetDefaultFeeOk(t *testing.T) {
	glb := New(t, false, false)
	chain := glb.NewChain(nil, "chain1")

	req := NewCall(root.Interface.Name, root.FuncSetDefaultFee,
		root.ParamOwnerFee, 1000,
		root.ParamValidatorFee, 499,
	)
	_, err := chain.PostRequest(req, nil)
	require.NoError(t, err)
	checkFees(chain, root.Interface.Name, 1000, 499)
	checkFees(chain, accountsc.Interface.Name, 1000, 499)
	checkFees(chain, blob.Interface.Name, 1000, 499)
}

func TestSetDefaultFeeFailNegative1(t *testing.T) {
	glb := New(t, false, false)
	chain := glb.NewChain(nil, "chain1")

	req := NewCall(root.Interface.Name, root.FuncSetDefaultFee, root.ParamOwnerFee, -2)
	_, err := chain.PostRequest(req, nil)
	require.Error(t, err)

	checkFees(chain, root.Interface.Name, 0, 0)
	checkFees(chain, accountsc.Interface.Name, 0, 0)
	checkFees(chain, blob.Interface.Name, 0, 0)
}

func TestSetDefaultFeeFailNegative2(t *testing.T) {
	glb := New(t, false, false)
	chain := glb.NewChain(nil, "chain1")

	req := NewCall(root.Interface.Name, root.FuncSetDefaultFee, root.ParamValidatorFee, -100)
	_, err := chain.PostRequest(req, nil)
	require.Error(t, err)

	checkFees(chain, root.Interface.Name, 0, 0)
	checkFees(chain, accountsc.Interface.Name, 0, 0)
	checkFees(chain, blob.Interface.Name, 0, 0)
}

func TestSetContractValidatorFeeOk(t *testing.T) {
	glb := New(t, false, false)
	chain := glb.NewChain(nil, "chain1")

	req := NewCall(root.Interface.Name, root.FuncSetContractFee,
		root.ParamHname, blob.Interface.Hname(),
		root.ParamValidatorFee, 1000,
	)
	_, err := chain.PostRequest(req, nil)
	require.NoError(t, err)

	checkFees(chain, root.Interface.Name, 0, 0)
	checkFees(chain, accountsc.Interface.Name, 0, 0)
	checkFees(chain, blob.Interface.Name, 0, 1000)
}

func TestSetContractOwnerFeeOk(t *testing.T) {
	glb := New(t, false, false)
	chain := glb.NewChain(nil, "chain1")

	req := NewCall(root.Interface.Name, root.FuncSetContractFee,
		root.ParamHname, accountsc.Interface.Hname(),
		root.ParamOwnerFee, 499,
	)
	_, err := chain.PostRequest(req, nil)
	require.NoError(t, err)

	checkFees(chain, root.Interface.Name, 0, 0)
	checkFees(chain, accountsc.Interface.Name, 499, 0)
	checkFees(chain, blob.Interface.Name, 0, 0)
}

func TestSetContractFeeWithDefault(t *testing.T) {
	glb := New(t, false, false)
	chain := glb.NewChain(nil, "chain1")

	req := NewCall(root.Interface.Name, root.FuncSetContractFee,
		root.ParamHname, blob.Interface.Hname(),
		root.ParamValidatorFee, 1000,
	)
	_, err := chain.PostRequest(req, nil)
	require.NoError(t, err)

	checkFees(chain, root.Interface.Name, 0, 0)
	checkFees(chain, accountsc.Interface.Name, 0, 0)
	checkFees(chain, blob.Interface.Name, 0, 1000)

	req = NewCall(root.Interface.Name, root.FuncSetDefaultFee,
		root.ParamOwnerFee, 499,
	)
	_, err = chain.PostRequest(req, nil)
	require.NoError(t, err)

	checkFees(chain, root.Interface.Name, 499, 0)
	checkFees(chain, accountsc.Interface.Name, 499, 0)
	checkFees(chain, blob.Interface.Name, 499, 1000)

	req = NewCall(root.Interface.Name, root.FuncSetDefaultFee, root.ParamValidatorFee, 1999)
	//.WithTransfer(
	//		map[balance.Color]int64{
	//			balance.ColorIOTA: 800,
	//		},
	//	)
	_, err = chain.PostRequest(req, nil)
	require.NoError(t, err)

	checkFees(chain, root.Interface.Name, 499, 1999)
	checkFees(chain, accountsc.Interface.Name, 499, 1999)
	checkFees(chain, blob.Interface.Name, 499, 1000)
}

func TestFeeNotEnough(t *testing.T) {
	glb := New(t, false, false)
	chain := glb.NewChain(nil, "chain1")

	req := NewCall(root.Interface.Name, root.FuncSetContractFee,
		root.ParamHname, root.Interface.Hname(),
		root.ParamValidatorFee, 499,
	)
	_, err := chain.PostRequest(req, nil)
	require.NoError(t, err)

	checkFees(chain, root.Interface.Name, 0, 499)
	checkFees(chain, accountsc.Interface.Name, 0, 0)
	checkFees(chain, blob.Interface.Name, 0, 0)

	user := glb.NewSigSchemeWithFunds()
	req = NewCall(root.Interface.Name, root.FuncSetDefaultFee,
		root.ParamOwnerFee, 1000,
	)
	_, err = chain.PostRequest(req, user)
	require.Error(t, err)

	checkFees(chain, root.Interface.Name, 0, 499)
	checkFees(chain, accountsc.Interface.Name, 0, 0)
	checkFees(chain, blob.Interface.Name, 0, 0)
}

func TestFeeOwnerDontNeed(t *testing.T) {
	glb := New(t, false, false)
	chain := glb.NewChain(nil, "chain1")

	req := NewCall(root.Interface.Name, root.FuncSetContractFee,
		root.ParamHname, root.Interface.Hname(),
		root.ParamValidatorFee, 499,
	)
	_, err := chain.PostRequest(req, nil)
	require.NoError(t, err)

	checkFees(chain, root.Interface.Name, 0, 499)
	checkFees(chain, accountsc.Interface.Name, 0, 0)
	checkFees(chain, blob.Interface.Name, 0, 0)

	req = NewCall(root.Interface.Name, root.FuncSetDefaultFee,
		root.ParamOwnerFee, 1000,
	)
	_, err = chain.PostRequest(req, nil)
	require.NoError(t, err)

	checkFees(chain, root.Interface.Name, 1000, 499)
	checkFees(chain, accountsc.Interface.Name, 1000, 0)
	checkFees(chain, blob.Interface.Name, 1000, 0)
}
