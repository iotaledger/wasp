// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package testcore

import (
	"testing"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/stretchr/testify/require"
)

func checkFees(chain *solo.Chain, contract string, expectedOf, expectedVf uint64) {
	col, ownerFee, validatorFee := chain.GetFeeInfo(contract)
	require.EqualValues(chain.Env.T, ledgerstate.ColorIOTA, col)
	require.EqualValues(chain.Env.T, int(expectedOf), int(ownerFee))
	require.EqualValues(chain.Env.T, int(expectedVf), int(validatorFee))
}

func TestFeeBasic(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")
	checkFees(chain, root.Interface.Name, 0, 0)
	checkFees(chain, accounts.Interface.Name, 0, 0)
	checkFees(chain, blob.Interface.Name, 0, 0)

	chain.AssertOwnersIotas(1)
	chain.AssertTotalIotas(1)
}

func TestSetDefaultFeeNotAuthorized(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	user, _ := env.NewKeyPairWithFunds()

	req := solo.NewCallParams(root.Interface.Name, root.FuncSetDefaultFee, root.ParamOwnerFee, 1000)
	_, err := chain.PostRequestSync(req, user)
	require.Error(t, err)

	checkFees(chain, root.Interface.Name, 0, 0)
	checkFees(chain, accounts.Interface.Name, 0, 0)
	checkFees(chain, blob.Interface.Name, 0, 0)

	chain.AssertOwnersIotas(1)
	chain.AssertTotalIotas(1)
}

func TestSetContractFeeNotAuthorized(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	user, _ := env.NewKeyPairWithFunds()

	req := solo.NewCallParams(root.Interface.Name, root.FuncSetContractFee, root.ParamOwnerFee, 1000)
	_, err := chain.PostRequestSync(req, user)
	require.Error(t, err)

	checkFees(chain, root.Interface.Name, 0, 0)
	checkFees(chain, accounts.Interface.Name, 0, 0)
	checkFees(chain, blob.Interface.Name, 0, 0)

	chain.AssertOwnersIotas(1)
	chain.AssertTotalIotas(1)
}

func TestSetDefaultOwnerFeeOk(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	req := solo.NewCallParams(root.Interface.Name, root.FuncSetDefaultFee,
		root.ParamOwnerFee, 1000,
	).WithIotas(1)
	_, err := chain.PostRequestSync(req, nil)
	require.NoError(t, err)
	checkFees(chain, root.Interface.Name, 1000, 0)
	checkFees(chain, accounts.Interface.Name, 1000, 0)
	checkFees(chain, blob.Interface.Name, 1000, 0)

	chain.AssertOwnersIotas(2)
	chain.AssertTotalIotas(2)
}

func TestSetDefaultValidatorFeeOk(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	req := solo.NewCallParams(root.Interface.Name, root.FuncSetDefaultFee,
		root.ParamValidatorFee, 499,
	).WithIotas(1)
	_, err := chain.PostRequestSync(req, nil)
	require.NoError(t, err)
	checkFees(chain, root.Interface.Name, 0, 499)
	checkFees(chain, accounts.Interface.Name, 0, 499)
	checkFees(chain, blob.Interface.Name, 0, 499)

	chain.AssertOwnersIotas(2)
	chain.AssertTotalIotas(2)
}

func TestSetDefaultFeeOk(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	req := solo.NewCallParams(root.Interface.Name, root.FuncSetDefaultFee,
		root.ParamOwnerFee, 1000,
		root.ParamValidatorFee, 499,
	).WithIotas(1)
	_, err := chain.PostRequestSync(req, nil)
	require.NoError(t, err)
	checkFees(chain, root.Interface.Name, 1000, 499)
	checkFees(chain, accounts.Interface.Name, 1000, 499)
	checkFees(chain, blob.Interface.Name, 1000, 499)

	chain.AssertOwnersIotas(2)
	chain.AssertTotalIotas(2)
}

func TestSetDefaultFeeFailNegative1(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	req := solo.NewCallParams(root.Interface.Name, root.FuncSetDefaultFee, root.ParamOwnerFee, -2).WithIotas(1)
	_, err := chain.PostRequestSync(req, nil)
	require.Error(t, err)

	checkFees(chain, root.Interface.Name, 0, 0)
	checkFees(chain, accounts.Interface.Name, 0, 0)
	checkFees(chain, blob.Interface.Name, 0, 0)

	chain.AssertOwnersIotas(1)
	chain.AssertTotalIotas(1)
}

func TestSetDefaultFeeFailNegative2(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	req := solo.NewCallParams(root.Interface.Name, root.FuncSetDefaultFee, root.ParamValidatorFee, -100).WithIotas(1)
	_, err := chain.PostRequestSync(req, nil)
	require.Error(t, err)

	checkFees(chain, root.Interface.Name, 0, 0)
	checkFees(chain, accounts.Interface.Name, 0, 0)
	checkFees(chain, blob.Interface.Name, 0, 0)

	chain.AssertOwnersIotas(1)
	chain.AssertTotalIotas(1)
}

func TestSetContractValidatorFeeOk(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	req := solo.NewCallParams(root.Interface.Name, root.FuncSetContractFee,
		root.ParamHname, blob.Interface.Hname(),
		root.ParamValidatorFee, 1000,
	).WithIotas(1)
	_, err := chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	checkFees(chain, root.Interface.Name, 0, 0)
	checkFees(chain, accounts.Interface.Name, 0, 0)
	checkFees(chain, blob.Interface.Name, 0, 1000)

	chain.AssertOwnersIotas(2)
	chain.AssertTotalIotas(2)
}

func TestSetContractOwnerFeeOk(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	req := solo.NewCallParams(root.Interface.Name, root.FuncSetContractFee,
		root.ParamHname, accounts.Interface.Hname(),
		root.ParamOwnerFee, 499,
	).WithIotas(1)
	_, err := chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	checkFees(chain, root.Interface.Name, 0, 0)
	checkFees(chain, accounts.Interface.Name, 499, 0)
	checkFees(chain, blob.Interface.Name, 0, 0)

	chain.AssertOwnersIotas(2)
	chain.AssertTotalIotas(2)
}

func TestSetContractFeeWithDefault(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	req := solo.NewCallParams(root.Interface.Name, root.FuncSetContractFee,
		root.ParamHname, blob.Interface.Hname(),
		root.ParamValidatorFee, 1000,
	).WithIotas(1)
	_, err := chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	checkFees(chain, root.Interface.Name, 0, 0)
	checkFees(chain, accounts.Interface.Name, 0, 0)
	checkFees(chain, blob.Interface.Name, 0, 1000)

	req = solo.NewCallParams(root.Interface.Name, root.FuncSetDefaultFee,
		root.ParamOwnerFee, 499,
	).WithIotas(1)
	_, err = chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	checkFees(chain, root.Interface.Name, 499, 0)
	checkFees(chain, accounts.Interface.Name, 499, 0)
	checkFees(chain, blob.Interface.Name, 499, 1000)

	req = solo.NewCallParams(root.Interface.Name, root.FuncSetDefaultFee, root.ParamValidatorFee, 1999).WithIotas(1)
	//.WithTransfers(
	//		map[ledgerstate.Color]uint64{
	//			ledgerstate.ColorIOTA: 800,
	//		},
	//	)
	_, err = chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	checkFees(chain, root.Interface.Name, 499, 1999)
	checkFees(chain, accounts.Interface.Name, 499, 1999)
	checkFees(chain, blob.Interface.Name, 499, 1000)

	chain.AssertOwnersIotas(4)
	chain.AssertTotalIotas(4)
}

func TestFeeNotEnough(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	req := solo.NewCallParams(root.Interface.Name, root.FuncSetContractFee,
		root.ParamHname, root.Interface.Hname(),
		root.ParamValidatorFee, 499,
	).WithIotas(1)
	_, err := chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	checkFees(chain, root.Interface.Name, 0, 499)
	checkFees(chain, accounts.Interface.Name, 0, 0)
	checkFees(chain, blob.Interface.Name, 0, 0)

	chain.AssertOwnersIotas(2)
	chain.AssertTotalIotas(2)

	user, _ := env.NewKeyPairWithFunds()
	req = solo.NewCallParams(root.Interface.Name, root.FuncSetDefaultFee,
		root.ParamOwnerFee, 1000,
	).WithIotas(99)
	_, err = chain.PostRequestSync(req, user)
	require.Error(t, err)

	checkFees(chain, root.Interface.Name, 0, 499)
	checkFees(chain, accounts.Interface.Name, 0, 0)
	checkFees(chain, blob.Interface.Name, 0, 0)

	// TODO no validator was provided, so iotas end up in null account
	chain.AssertIotas(&coretypes.NilAgentID, 99)
	chain.AssertOwnersIotas(2)
	chain.AssertTotalIotas(101)
}

func TestFeeOwnerDontNeed(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	req := solo.NewCallParams(root.Interface.Name, root.FuncSetContractFee,
		root.ParamHname, root.Interface.Hname(),
		root.ParamValidatorFee, 499,
	).WithIotas(1)
	_, err := chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	checkFees(chain, root.Interface.Name, 0, 499)
	checkFees(chain, accounts.Interface.Name, 0, 0)
	checkFees(chain, blob.Interface.Name, 0, 0)

	req = solo.NewCallParams(root.Interface.Name, root.FuncSetDefaultFee,
		root.ParamOwnerFee, 1000,
	).WithIotas(99)
	_, err = chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	checkFees(chain, root.Interface.Name, 1000, 499)
	checkFees(chain, accounts.Interface.Name, 1000, 0)
	checkFees(chain, blob.Interface.Name, 1000, 0)

	chain.AssertOwnersIotas(101)
	chain.AssertTotalIotas(101)
}
