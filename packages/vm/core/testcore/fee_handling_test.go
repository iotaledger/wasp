// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package testcore

import (
	"testing"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	chain.AssertIotas(&chain.OriginatorAgentID, 0)
	chain.AssertCommonAccountIotas(1)
	env.AssertAddressBalance(chain.OriginatorAddress, ledgerstate.ColorIOTA, solo.Saldo-solo.ChainDustThreshold-1)
	chain.AssertTotalIotas(1)
	chain.AssertCommonAccountIotas(1)

	checkFees(chain, blob.Contract.Name, 0, 0)
	checkFees(chain, root.Contract.Name, 0, 0)
	checkFees(chain, accounts.Contract.Name, 0, 0)
}

func TestBase(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	req := solo.NewCallParams(root.Contract.Name, root.FuncSetContractFee.Name,
		root.ParamHname, blob.Contract.Hname(),
		root.ParamOwnerFee, 5,
	).WithIotas(1)
	_, err := chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	chain.AssertCommonAccountIotas(2)
	chain.AssertTotalIotas(2)
	env.AssertAddressBalance(chain.OriginatorAddress, ledgerstate.ColorIOTA, solo.Saldo-solo.ChainDustThreshold-2)

	checkFees(chain, blob.Contract.Name, 5, 0)
}

//nolint:dupl
func TestFeeIsEnough1(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	req := solo.NewCallParams(root.Contract.Name, root.FuncSetContractFee.Name,
		root.ParamHname, blob.Contract.Hname(),
		root.ParamOwnerFee, 1,
	).WithIotas(1)
	_, err := chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	chain.AssertCommonAccountIotas(2)
	chain.AssertTotalIotas(2)
	env.AssertAddressBalance(chain.OriginatorAddress, ledgerstate.ColorIOTA, solo.Saldo-solo.ChainDustThreshold-2)

	checkFees(chain, blob.Contract.Name, 1, 0)

	// the upload blob takes fees itself
	_, err = chain.UploadBlob(nil,
		blob.VarFieldVMType, "dummyType",
		blob.VarFieldProgramBinary, "dummyBinary",
	)
	require.NoError(t, err)

	chain.AssertCommonAccountIotas(2 + 1)
	chain.AssertTotalIotas(2 + 1)
	chain.AssertAccountBalance(&chain.OriginatorAgentID, ledgerstate.ColorIOTA, 0)
	env.AssertAddressBalance(chain.OriginatorAddress, ledgerstate.ColorIOTA, solo.Saldo-solo.ChainDustThreshold-2-1)
}

//nolint:dupl
func TestFeeIsEnough2(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	req := solo.NewCallParams(root.Contract.Name, root.FuncSetContractFee.Name,
		root.ParamHname, blob.Contract.Hname(),
		root.ParamOwnerFee, 10,
	).WithIotas(1)
	_, err := chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	chain.AssertCommonAccountIotas(2)
	chain.AssertTotalIotas(2)
	env.AssertAddressBalance(chain.OriginatorAddress, ledgerstate.ColorIOTA, solo.Saldo-solo.ChainDustThreshold-2)

	checkFees(chain, blob.Contract.Name, 10, 0)

	// the upload blob takes fees itself
	_, err = chain.UploadBlob(nil,
		blob.VarFieldVMType, "dummyType",
		blob.VarFieldProgramBinary, "dummyBinary",
	)
	require.NoError(t, err)

	chain.AssertCommonAccountIotas(2 + 10)
	chain.AssertTotalIotas(2 + 10)
	chain.AssertAccountBalance(&chain.OriginatorAgentID, ledgerstate.ColorIOTA, 0)
	env.AssertAddressBalance(chain.OriginatorAddress, ledgerstate.ColorIOTA, solo.Saldo-solo.ChainDustThreshold-2-10)
}

func TestFeesNoNeed(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	req := solo.NewCallParams(root.Contract.Name, root.FuncSetContractFee.Name,
		root.ParamHname, blob.Contract.Hname(),
		root.ParamOwnerFee, 10,
	).WithIotas(1)
	_, err := chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	chain.AssertCommonAccountIotas(2)
	chain.AssertTotalIotas(2)
	env.AssertAddressBalance(chain.OriginatorAddress, ledgerstate.ColorIOTA, solo.Saldo-solo.ChainDustThreshold-2)

	checkFees(chain, blob.Contract.Name, 10, 0)

	req = solo.NewCallParams(blob.Contract.Name, blob.FuncStoreBlob.Name, "par1", []byte("data1"))
	req.WithIotas(7)
	_, err = chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	chain.AssertCommonAccountIotas(2 + 7)
	chain.AssertTotalIotas(2 + 7)
	chain.AssertAccountBalance(&chain.OriginatorAgentID, ledgerstate.ColorIOTA, 0)
	env.AssertAddressBalance(chain.OriginatorAddress, ledgerstate.ColorIOTA, solo.Saldo-solo.ChainDustThreshold-2-7)
}

func TestFeesNotEnough(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	user, userAddr := env.NewKeyPairWithFunds()
	userAgentID := iscp.NewAgentID(userAddr, 0)

	req := solo.NewCallParams(root.Contract.Name, root.FuncSetContractFee.Name,
		root.ParamHname, blob.Contract.Hname(),
		root.ParamOwnerFee, 10,
	).WithIotas(1)
	_, err := chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	checkFees(chain, blob.Contract.Name, 10, 0)

	chain.AssertCommonAccountIotas(2)
	chain.AssertTotalIotas(2)
	chain.AssertIotas(userAgentID, 0)
	env.AssertAddressIotas(userAddr, solo.Saldo)

	req = solo.NewCallParams(blob.Contract.Name, blob.FuncStoreBlob.Name, "par1", []byte("data1"))
	req.WithIotas(7)
	_, err = chain.PostRequestSync(req, user)
	require.Error(t, err)

	chain.AssertCommonAccountIotas(2 + 7)
	chain.AssertTotalIotas(2 + 7)
	chain.AssertIotas(userAgentID, 0)
	env.AssertAddressIotas(userAddr, solo.Saldo-7)
}
