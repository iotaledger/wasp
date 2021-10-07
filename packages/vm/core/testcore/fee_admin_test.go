// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package testcore

import (
	"testing"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/colored"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/stretchr/testify/require"
)

func checkFees(chain *solo.Chain, contract string, expectedOf, expectedVf uint64) {
	col, ownerFee, validatorFee := chain.GetFeeInfo(contract)
	require.EqualValues(chain.Env.T, colored.IOTA, col)
	require.EqualValues(chain.Env.T, int(expectedOf), int(ownerFee))
	require.EqualValues(chain.Env.T, int(expectedVf), int(validatorFee))
}

func TestFeeBasic(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")
	checkFees(chain, root.Contract.Name, 0, 0)
	checkFees(chain, accounts.Contract.Name, 0, 0)
	checkFees(chain, blob.Contract.Name, 0, 0)

	chain.AssertCommonAccountIotas(1)
	chain.AssertTotalIotas(1)
}

func TestSetDefaultFeeNotAuthorized(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	user, _ := env.NewKeyPairWithFunds()

	req := solo.NewCallParams(governance.Contract.Name, governance.FuncSetChainInfo.Name, governance.ParamOwnerFee, 1000)
	_, err := chain.PostRequestSync(req, user)
	require.Error(t, err)

	checkFees(chain, root.Contract.Name, 0, 0)
	checkFees(chain, accounts.Contract.Name, 0, 0)
	checkFees(chain, blob.Contract.Name, 0, 0)

	chain.AssertCommonAccountIotas(1)
	chain.AssertTotalIotas(1)
}

func TestSetContractFeeNotAuthorized(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	user, _ := env.NewKeyPairWithFunds()

	req := solo.NewCallParams(governance.Contract.Name, governance.FuncSetContractFee.Name, governance.ParamOwnerFee, 1000)
	_, err := chain.PostRequestSync(req, user)
	require.Error(t, err)

	checkFees(chain, root.Contract.Name, 0, 0)
	checkFees(chain, accounts.Contract.Name, 0, 0)
	checkFees(chain, blob.Contract.Name, 0, 0)

	chain.AssertCommonAccountIotas(1)
	chain.AssertTotalIotas(1)
}

func TestSetDefaultOwnerFeeOk(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	req := solo.NewCallParams(governance.Contract.Name, governance.FuncSetChainInfo.Name,
		governance.ParamOwnerFee, 1000,
	)
	_, err := chain.PostRequestSync(req.WithIotas(1), nil)
	require.NoError(t, err)
	checkFees(chain, root.Contract.Name, 1000, 0)
	checkFees(chain, accounts.Contract.Name, 1000, 0)
	checkFees(chain, blob.Contract.Name, 1000, 0)

	chain.AssertCommonAccountIotas(2)
	chain.AssertTotalIotas(2)
}

func TestSetDefaultValidatorFeeOk(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	req := solo.NewCallParams(governance.Contract.Name, governance.FuncSetChainInfo.Name,
		governance.ParamValidatorFee, 499,
	)
	_, err := chain.PostRequestSync(req.WithIotas(1), nil)
	require.NoError(t, err)
	checkFees(chain, root.Contract.Name, 0, 499)
	checkFees(chain, accounts.Contract.Name, 0, 499)
	checkFees(chain, blob.Contract.Name, 0, 499)

	chain.AssertCommonAccountIotas(2)
	chain.AssertTotalIotas(2)
}

func TestSetDefaultFeeOk(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	req := solo.NewCallParams(governance.Contract.Name, governance.FuncSetChainInfo.Name,
		governance.ParamOwnerFee, 1000,
		governance.ParamValidatorFee, 499,
	)
	_, err := chain.PostRequestSync(req.WithIotas(1), nil)
	require.NoError(t, err)
	checkFees(chain, root.Contract.Name, 1000, 499)
	checkFees(chain, accounts.Contract.Name, 1000, 499)
	checkFees(chain, blob.Contract.Name, 1000, 499)

	chain.AssertCommonAccountIotas(2)
	chain.AssertTotalIotas(2)
}

func TestSetDefaultFeeFailNegative1(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	req := solo.NewCallParams(governance.Contract.Name, governance.FuncSetChainInfo.Name, governance.ParamOwnerFee, -2)
	_, err := chain.PostRequestSync(req.WithIotas(1), nil)
	require.NoError(t, err)

	checkFees(chain, root.Contract.Name, 0, 0)
	checkFees(chain, accounts.Contract.Name, 0, 0)
	checkFees(chain, blob.Contract.Name, 0, 0)

	chain.AssertCommonAccountIotas(2)
	chain.AssertTotalIotas(2)
}

func TestSetDefaultFeeFailNegative2(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	req := solo.NewCallParams(governance.Contract.Name, governance.FuncSetChainInfo.Name, governance.ParamValidatorFee, -100)
	_, err := chain.PostRequestSync(req.WithIotas(1), nil)
	require.NoError(t, err)

	checkFees(chain, root.Contract.Name, 0, 0)
	checkFees(chain, accounts.Contract.Name, 0, 0)
	checkFees(chain, blob.Contract.Name, 0, 0)

	chain.AssertCommonAccountIotas(2)
	chain.AssertTotalIotas(2)
}

func TestSetContractValidatorFeeOk(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	req := solo.NewCallParams(governance.Contract.Name, governance.FuncSetContractFee.Name,
		governance.ParamHname, blob.Contract.Hname(),
		governance.ParamValidatorFee, 1000,
	)
	_, err := chain.PostRequestSync(req.WithIotas(1), nil)
	require.NoError(t, err)

	checkFees(chain, root.Contract.Name, 0, 0)
	checkFees(chain, accounts.Contract.Name, 0, 0)
	checkFees(chain, blob.Contract.Name, 0, 1000)

	chain.AssertCommonAccountIotas(2)
	chain.AssertTotalIotas(2)
}

func TestSetContractOwnerFeeOk(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	req := solo.NewCallParams(governance.Contract.Name, governance.FuncSetContractFee.Name,
		governance.ParamHname, accounts.Contract.Hname(),
		governance.ParamOwnerFee, 499,
	)
	_, err := chain.PostRequestSync(req.WithIotas(1), nil)
	require.NoError(t, err)

	checkFees(chain, root.Contract.Name, 0, 0)
	checkFees(chain, accounts.Contract.Name, 499, 0)
	checkFees(chain, blob.Contract.Name, 0, 0)

	chain.AssertCommonAccountIotas(2)
	chain.AssertTotalIotas(2)
}

func TestSetContractFeeWithDefault(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	req := solo.NewCallParams(governance.Contract.Name, governance.FuncSetContractFee.Name,
		governance.ParamHname, blob.Contract.Hname(),
		governance.ParamValidatorFee, 1000,
	)
	_, err := chain.PostRequestSync(req.WithIotas(1), nil)
	require.NoError(t, err)

	checkFees(chain, root.Contract.Name, 0, 0)
	checkFees(chain, accounts.Contract.Name, 0, 0)
	checkFees(chain, blob.Contract.Name, 0, 1000)

	req = solo.NewCallParams(governance.Contract.Name, governance.FuncSetChainInfo.Name,
		governance.ParamOwnerFee, 499,
	).WithIotas(1)
	_, err = chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	checkFees(chain, root.Contract.Name, 499, 0)
	checkFees(chain, accounts.Contract.Name, 499, 0)
	checkFees(chain, blob.Contract.Name, 499, 1000)

	req = solo.NewCallParams(governance.Contract.Name, governance.FuncSetChainInfo.Name, governance.ParamValidatorFee, 1999).WithIotas(1)
	//.WithTransfers(
	//		map[ledgerstate.Color]uint64{
	//			ledgerstate.ColorIOTA: 800,
	//		},
	//	)
	_, err = chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	checkFees(chain, root.Contract.Name, 499, 1999)
	checkFees(chain, accounts.Contract.Name, 499, 1999)
	checkFees(chain, blob.Contract.Name, 499, 1000)

	chain.AssertCommonAccountIotas(4)
	chain.AssertTotalIotas(4)
}

func TestFeeNotEnough(t *testing.T) {
	env := solo.New(t, false, false)
	_, validatorFeeTargetAddr := env.NewKeyPair()
	validatorFeeTargetAgentID := iscp.NewAgentID(validatorFeeTargetAddr, 0)

	chain := env.NewChain(nil, "chain1", validatorFeeTargetAgentID)
	req := solo.NewCallParams(governance.Contract.Name, governance.FuncSetContractFee.Name,
		governance.ParamHname, governance.Contract.Hname(),
		governance.ParamValidatorFee, 499,
	)
	_, err := chain.PostRequestSync(req.WithIotas(1), nil)
	require.NoError(t, err)

	checkFees(chain, governance.Contract.Name, 0, 499)
	checkFees(chain, accounts.Contract.Name, 0, 0)
	checkFees(chain, blob.Contract.Name, 0, 0)

	chain.AssertCommonAccountIotas(2)
	chain.AssertTotalIotas(2)

	user, _ := env.NewKeyPairWithFunds()
	req = solo.NewCallParams(governance.Contract.Name, governance.FuncSetChainInfo.Name,
		governance.ParamOwnerFee, 1000,
	).WithIotas(99)
	_, err = chain.PostRequestSync(req, user)
	require.Error(t, err)

	checkFees(chain, governance.Contract.Name, 0, 499)
	checkFees(chain, accounts.Contract.Name, 0, 0)
	checkFees(chain, blob.Contract.Name, 0, 0)

	chain.AssertCommonAccountIotas(2)
	chain.AssertTotalIotas(101)
	chain.AssertIotas(validatorFeeTargetAgentID, 99)
}

func TestFeeOwnerDontNeed(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	req := solo.NewCallParams(governance.Contract.Name, governance.FuncSetContractFee.Name,
		governance.ParamHname, governance.Contract.Hname(),
		governance.ParamValidatorFee, 499,
	)
	_, err := chain.PostRequestSync(req.WithIotas(1), nil)
	require.NoError(t, err)

	checkFees(chain, governance.Contract.Name, 0, 499)
	checkFees(chain, accounts.Contract.Name, 0, 0)
	checkFees(chain, blob.Contract.Name, 0, 0)

	req = solo.NewCallParams(governance.Contract.Name, governance.FuncSetChainInfo.Name,
		governance.ParamOwnerFee, 1000,
	).WithIotas(99)
	_, err = chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	checkFees(chain, governance.Contract.Name, 1000, 499)
	checkFees(chain, accounts.Contract.Name, 1000, 0)
	checkFees(chain, blob.Contract.Name, 1000, 0)

	chain.AssertCommonAccountIotas(101)
	chain.AssertTotalIotas(101)
}

func TestRevertContractFeeToZero(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	req := solo.NewCallParams(governance.Contract.Name, governance.FuncSetContractFee.Name,
		governance.ParamHname, blob.Contract.Hname(),
		governance.ParamValidatorFee, 1000,
		governance.ParamOwnerFee, 1000,
	)
	_, err := chain.PostRequestSync(req.WithIotas(1), nil)
	require.NoError(t, err)

	checkFees(chain, root.Contract.Name, 0, 0)
	checkFees(chain, accounts.Contract.Name, 0, 0)
	checkFees(chain, blob.Contract.Name, 1000, 1000)

	req = solo.NewCallParams(governance.Contract.Name, governance.FuncSetContractFee.Name,
		governance.ParamHname, blob.Contract.Hname(),
		governance.ParamValidatorFee, 0,
		governance.ParamOwnerFee, 0,
	).WithIotas(1)
	_, err = chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	checkFees(chain, blob.Contract.Name, 0, 0)

	chain.AssertCommonAccountIotas(3)
	chain.AssertTotalIotas(3)
}
