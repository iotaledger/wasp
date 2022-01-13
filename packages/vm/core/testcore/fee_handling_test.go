// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package testcore

import (
	"testing"


	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	chain.AssertL2AccountIotas(chain.OriginatorAgentID, 0)
	chain.AssertCommonAccountIotas(1)
	env.AssertAddressNativeTokenBalance(chain.OriginatorAddress, colored.IOTA, solo.Saldo-solo.ChainDustThreshold-1)
	chain.AssertL2TotalIotas(1)
	chain.AssertCommonAccountIotas(1)

	checkFees(chain, blob.Contract.Name, 0, 0)
	checkFees(chain, root.Contract.Name, 0, 0)
	checkFees(chain, accounts.Contract.Name, 0, 0)
}

func TestBase(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	req := solo.NewCallParams(governance.Contract.Name, governance.FuncSetContractFee.Name,
		governance.ParamHname, blob.Contract.Hname(),
		governance.ParamOwnerFee, 5,
	)
	_, err := chain.PostRequestSync(req.AddAssetsIotas(1), nil)
	require.NoError(t, err)

	chain.AssertCommonAccountIotas(2)
	chain.AssertL2TotalIotas(2)
	env.AssertAddressNativeTokenBalance(chain.OriginatorAddress, colored.IOTA, solo.Saldo-solo.ChainDustThreshold-2)

	checkFees(chain, blob.Contract.Name, 5, 0)
}

//nolint:dupl
func TestFeeIsEnough1(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	req := solo.NewCallParams(governance.Contract.Name, governance.FuncSetContractFee.Name,
		governance.ParamHname, blob.Contract.Hname(),
		governance.ParamOwnerFee, 1,
	)
	_, err := chain.PostRequestSync(req.AddAssetsIotas(1), nil)
	require.NoError(t, err)

	chain.AssertCommonAccountIotas(2)
	chain.AssertL2TotalIotas(2)
	env.AssertAddressNativeTokenBalance(chain.OriginatorAddress, colored.IOTA, solo.Saldo-solo.ChainDustThreshold-2)

	checkFees(chain, blob.Contract.Name, 1, 0)

	// the upload blob takes fees itself
	_, err = chain.UploadBlob(nil,
		blob.VarFieldVMType, "dummyType",
		blob.VarFieldProgramBinary, "dummyBinary",
	)
	require.NoError(t, err)

	chain.AssertCommonAccountIotas(2 + 1)
	chain.AssertL2TotalIotas(2 + 1)
	chain.AssertL2AccountNativeToken(chain.OriginatorAgentID, colored.IOTA, 0)
	env.AssertAddressNativeTokenBalance(chain.OriginatorAddress, colored.IOTA, solo.Saldo-solo.ChainDustThreshold-2-1)
}

//nolint:dupl
func TestFeeIsEnough2(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	req := solo.NewCallParams(governance.Contract.Name, governance.FuncSetContractFee.Name,
		governance.ParamHname, blob.Contract.Hname(),
		governance.ParamOwnerFee, 10,
	)
	_, err := chain.PostRequestSync(req.AddAssetsIotas(1), nil)
	require.NoError(t, err)

	chain.AssertCommonAccountIotas(2)
	chain.AssertL2TotalIotas(2)
	env.AssertAddressNativeTokenBalance(chain.OriginatorAddress, colored.IOTA, solo.Saldo-solo.ChainDustThreshold-2)

	checkFees(chain, blob.Contract.Name, 10, 0)

	// the upload blob takes fees itself
	_, err = chain.UploadBlob(nil,
		blob.VarFieldVMType, "dummyType",
		blob.VarFieldProgramBinary, "dummyBinary",
	)
	require.NoError(t, err)

	chain.AssertCommonAccountIotas(2 + 10)
	chain.AssertL2TotalIotas(2 + 10)
	chain.AssertL2AccountNativeToken(chain.OriginatorAgentID, colored.IOTA, 0)
	env.AssertAddressNativeTokenBalance(chain.OriginatorAddress, colored.IOTA, solo.Saldo-solo.ChainDustThreshold-2-10)
}

func TestFeesNoNeed(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	req := solo.NewCallParams(governance.Contract.Name, governance.FuncSetContractFee.Name,
		governance.ParamHname, blob.Contract.Hname(),
		governance.ParamOwnerFee, 10,
	)
	_, err := chain.PostRequestSync(req.AddAssetsIotas(1), nil)
	require.NoError(t, err)

	chain.AssertCommonAccountIotas(2)
	chain.AssertL2TotalIotas(2)
	env.AssertAddressNativeTokenBalance(chain.OriginatorAddress, colored.IOTA, solo.Saldo-solo.ChainDustThreshold-2)

	checkFees(chain, blob.Contract.Name, 10, 0)

	req = solo.NewCallParams(blob.Contract.Name, blob.FuncStoreBlob.Name, "par1", []byte("data1"))
	req.AddAssetsIotas(7)
	_, err = chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	chain.AssertCommonAccountIotas(2 + 7)
	chain.AssertL2TotalIotas(2 + 7)
	chain.AssertL2AccountNativeToken(chain.OriginatorAgentID, colored.IOTA, 0)
	env.AssertAddressNativeTokenBalance(chain.OriginatorAddress, colored.IOTA, solo.Saldo-solo.ChainDustThreshold-2-7)
}

func TestFeesNotEnough(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	user, userAddr := env.NewKeyPairWithFunds()
	userAgentID := iscp.NewAgentID(userAddr, 0)

	req := solo.NewCallParams(governance.Contract.Name, governance.FuncSetContractFee.Name,
		governance.ParamHname, blob.Contract.Hname(),
		governance.ParamOwnerFee, 10,
	)
	_, err := chain.PostRequestSync(req.AddAssetsIotas(1), nil)
	require.NoError(t, err)

	checkFees(chain, blob.Contract.Name, 10, 0)

	chain.AssertCommonAccountIotas(2)
	chain.AssertL2TotalIotas(2)
	chain.AssertL2AccountIotas(userAgentID, 0)
	env.AssertAddressIotas(userAddr, solo.Saldo)

	req = solo.NewCallParams(blob.Contract.Name, blob.FuncStoreBlob.Name, "par1", []byte("data1"))
	req.AddAssetsIotas(7)
	_, err = chain.PostRequestSync(req, user)
	require.Error(t, err)

	chain.AssertCommonAccountIotas(2 + 7)
	chain.AssertL2TotalIotas(2 + 7)
	chain.AssertL2AccountIotas(userAgentID, 0)
	env.AssertAddressIotas(userAddr, solo.Saldo-7)
}
