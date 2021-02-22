// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package testcore

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/stretchr/testify/require"
	"testing"
)

func checkFees(chain *solo.Chain, contract string, expectedOf, expectedVf int64) {
	col, ownerFee, validatorFee := chain.GetFeeInfo(contract)
	require.EqualValues(chain.Env.T, balance.ColorIOTA, col)
	require.EqualValues(chain.Env.T, expectedOf, ownerFee)
	require.EqualValues(chain.Env.T, expectedVf, validatorFee)
}

func TestFeeBasic(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")
	checkFees(chain, root.Interface.Name, 0, 0)
	checkFees(chain, accounts.Interface.Name, 0, 0)
	checkFees(chain, blob.Interface.Name, 0, 0)
}

func TestSetDefaultFeeNotAuthorized(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	user := env.NewSignatureSchemeWithFunds()

	req := solo.NewCallParams(root.Interface.Name, root.FuncSetDefaultFee, root.ParamOwnerFee, 1000)
	_, err := chain.PostRequestSync(req, user)
	require.Error(t, err)

	checkFees(chain, root.Interface.Name, 0, 0)
	checkFees(chain, accounts.Interface.Name, 0, 0)
	checkFees(chain, blob.Interface.Name, 0, 0)
}

func TestSetContractFeeNotAuthorized(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	user := env.NewSignatureSchemeWithFunds()

	req := solo.NewCallParams(root.Interface.Name, root.FuncSetContractFee, root.ParamOwnerFee, 1000)
	_, err := chain.PostRequestSync(req, user)
	require.Error(t, err)

	checkFees(chain, root.Interface.Name, 0, 0)
	checkFees(chain, accounts.Interface.Name, 0, 0)
	checkFees(chain, blob.Interface.Name, 0, 0)
}

func TestSetDefaultOwnerFeeOk(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	req := solo.NewCallParams(root.Interface.Name, root.FuncSetDefaultFee,
		root.ParamOwnerFee, 1000,
	)
	_, err := chain.PostRequestSync(req, nil)
	require.NoError(t, err)
	checkFees(chain, root.Interface.Name, 1000, 0)
	checkFees(chain, accounts.Interface.Name, 1000, 0)
	checkFees(chain, blob.Interface.Name, 1000, 0)
}

func TestSetDefaultValidatorFeeOk(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	req := solo.NewCallParams(root.Interface.Name, root.FuncSetDefaultFee,
		root.ParamValidatorFee, 499,
	)
	_, err := chain.PostRequestSync(req, nil)
	require.NoError(t, err)
	checkFees(chain, root.Interface.Name, 0, 499)
	checkFees(chain, accounts.Interface.Name, 0, 499)
	checkFees(chain, blob.Interface.Name, 0, 499)
}

func TestSetDefaultFeeOk(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	req := solo.NewCallParams(root.Interface.Name, root.FuncSetDefaultFee,
		root.ParamOwnerFee, 1000,
		root.ParamValidatorFee, 499,
	)
	_, err := chain.PostRequestSync(req, nil)
	require.NoError(t, err)
	checkFees(chain, root.Interface.Name, 1000, 499)
	checkFees(chain, accounts.Interface.Name, 1000, 499)
	checkFees(chain, blob.Interface.Name, 1000, 499)
}

func TestSetDefaultFeeFailNegative1(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	req := solo.NewCallParams(root.Interface.Name, root.FuncSetDefaultFee, root.ParamOwnerFee, -2)
	_, err := chain.PostRequestSync(req, nil)
	require.Error(t, err)

	checkFees(chain, root.Interface.Name, 0, 0)
	checkFees(chain, accounts.Interface.Name, 0, 0)
	checkFees(chain, blob.Interface.Name, 0, 0)
}

func TestSetDefaultFeeFailNegative2(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	req := solo.NewCallParams(root.Interface.Name, root.FuncSetDefaultFee, root.ParamValidatorFee, -100)
	_, err := chain.PostRequestSync(req, nil)
	require.Error(t, err)

	checkFees(chain, root.Interface.Name, 0, 0)
	checkFees(chain, accounts.Interface.Name, 0, 0)
	checkFees(chain, blob.Interface.Name, 0, 0)
}

func TestSetContractValidatorFeeOk(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	req := solo.NewCallParams(root.Interface.Name, root.FuncSetContractFee,
		root.ParamHname, blob.Interface.Hname(),
		root.ParamValidatorFee, 1000,
	)
	_, err := chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	checkFees(chain, root.Interface.Name, 0, 0)
	checkFees(chain, accounts.Interface.Name, 0, 0)
	checkFees(chain, blob.Interface.Name, 0, 1000)
}

func TestSetContractOwnerFeeOk(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	req := solo.NewCallParams(root.Interface.Name, root.FuncSetContractFee,
		root.ParamHname, accounts.Interface.Hname(),
		root.ParamOwnerFee, 499,
	)
	_, err := chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	checkFees(chain, root.Interface.Name, 0, 0)
	checkFees(chain, accounts.Interface.Name, 499, 0)
	checkFees(chain, blob.Interface.Name, 0, 0)
}

func TestSetContractFeeWithDefault(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	req := solo.NewCallParams(root.Interface.Name, root.FuncSetContractFee,
		root.ParamHname, blob.Interface.Hname(),
		root.ParamValidatorFee, 1000,
	)
	_, err := chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	checkFees(chain, root.Interface.Name, 0, 0)
	checkFees(chain, accounts.Interface.Name, 0, 0)
	checkFees(chain, blob.Interface.Name, 0, 1000)

	req = solo.NewCallParams(root.Interface.Name, root.FuncSetDefaultFee,
		root.ParamOwnerFee, 499,
	)
	_, err = chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	checkFees(chain, root.Interface.Name, 499, 0)
	checkFees(chain, accounts.Interface.Name, 499, 0)
	checkFees(chain, blob.Interface.Name, 499, 1000)

	req = solo.NewCallParams(root.Interface.Name, root.FuncSetDefaultFee, root.ParamValidatorFee, 1999)
	//.WithTransfers(
	//		map[balance.Color]int64{
	//			balance.ColorIOTA: 800,
	//		},
	//	)
	_, err = chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	checkFees(chain, root.Interface.Name, 499, 1999)
	checkFees(chain, accounts.Interface.Name, 499, 1999)
	checkFees(chain, blob.Interface.Name, 499, 1000)
}

func TestFeeNotEnough(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	req := solo.NewCallParams(root.Interface.Name, root.FuncSetContractFee,
		root.ParamHname, root.Interface.Hname(),
		root.ParamValidatorFee, 499,
	)
	_, err := chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	checkFees(chain, root.Interface.Name, 0, 499)
	checkFees(chain, accounts.Interface.Name, 0, 0)
	checkFees(chain, blob.Interface.Name, 0, 0)

	user := env.NewSignatureSchemeWithFunds()
	req = solo.NewCallParams(root.Interface.Name, root.FuncSetDefaultFee,
		root.ParamOwnerFee, 1000,
	)
	_, err = chain.PostRequestSync(req, user)
	require.Error(t, err)

	checkFees(chain, root.Interface.Name, 0, 499)
	checkFees(chain, accounts.Interface.Name, 0, 0)
	checkFees(chain, blob.Interface.Name, 0, 0)
}

func TestFeeOwnerDontNeed(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	req := solo.NewCallParams(root.Interface.Name, root.FuncSetContractFee,
		root.ParamHname, root.Interface.Hname(),
		root.ParamValidatorFee, 499,
	)
	_, err := chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	checkFees(chain, root.Interface.Name, 0, 499)
	checkFees(chain, accounts.Interface.Name, 0, 0)
	checkFees(chain, blob.Interface.Name, 0, 0)

	req = solo.NewCallParams(root.Interface.Name, root.FuncSetDefaultFee,
		root.ParamOwnerFee, 1000,
	)
	_, err = chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	checkFees(chain, root.Interface.Name, 1000, 499)
	checkFees(chain, accounts.Interface.Name, 1000, 0)
	checkFees(chain, blob.Interface.Name, 1000, 0)
}
