// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package testcore

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/blob"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestInit(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")

	chain.CheckAccountBalance(chain.OriginatorAgentID, balance.ColorIOTA, 1)
	glb.CheckUtxodbBalance(chain.OriginatorAddress, balance.ColorIOTA, testutil.RequestFundsAmount-2)
}

func TestBase(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")

	req := solo.NewCall(root.Interface.Name, root.FuncSetContractFee,
		root.ParamHname, blob.Interface.Hname(),
		root.ParamOwnerFee, 1,
	)
	_, err := chain.PostRequest(req, nil)
	require.NoError(t, err)

	chain.CheckAccountBalance(chain.OriginatorAgentID, balance.ColorIOTA, 2)
	glb.CheckUtxodbBalance(chain.OriginatorAddress, balance.ColorIOTA, testutil.RequestFundsAmount-3)

	checkFees(chain, blob.Interface.Name, 1, 0)
}

func TestFeeIsEnough1(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")

	req := solo.NewCall(root.Interface.Name, root.FuncSetContractFee,
		root.ParamHname, blob.Interface.Hname(),
		root.ParamOwnerFee, 1,
	)
	_, err := chain.PostRequest(req, nil)
	require.NoError(t, err)

	chain.CheckAccountBalance(chain.OriginatorAgentID, balance.ColorIOTA, 2)
	glb.CheckUtxodbBalance(chain.OriginatorAddress, balance.ColorIOTA, testutil.RequestFundsAmount-3)

	checkFees(chain, blob.Interface.Name, 1, 0)

	_, err = chain.UploadBlob(nil,
		blob.VarFieldVMType, "dummyType",
		blob.VarFieldProgramBinary, "dummyBinary",
	)
	require.NoError(t, err)

	chain.CheckAccountBalance(chain.OriginatorAgentID, balance.ColorIOTA, 3)
	glb.CheckUtxodbBalance(chain.OriginatorAddress, balance.ColorIOTA, testutil.RequestFundsAmount-5)
}

func TestFeeIsEnough2(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")

	req := solo.NewCall(root.Interface.Name, root.FuncSetContractFee,
		root.ParamHname, blob.Interface.Hname(),
		root.ParamOwnerFee, 2,
	)
	_, err := chain.PostRequest(req, nil)
	require.NoError(t, err)

	chain.CheckAccountBalance(chain.OriginatorAgentID, balance.ColorIOTA, 2)
	glb.CheckUtxodbBalance(chain.OriginatorAddress, balance.ColorIOTA, testutil.RequestFundsAmount-3)

	checkFees(chain, blob.Interface.Name, 2, 0)

	user := glb.NewSignatureSchemeWithFunds()
	userAgentID := coretypes.NewAgentIDFromAddress(user.Address())
	_, err = chain.UploadBlob(user,
		blob.VarFieldVMType, "dummyType",
		blob.VarFieldProgramBinary, "dummyBinary",
	)
	require.NoError(t, err)

	chain.CheckAccountBalance(chain.OriginatorAgentID, balance.ColorIOTA, 4)
	glb.CheckUtxodbBalance(chain.OriginatorAddress, balance.ColorIOTA, testutil.RequestFundsAmount-3)

	chain.CheckAccountBalance(userAgentID, balance.ColorIOTA, 1)
	glb.CheckUtxodbBalance(user.Address(), balance.ColorIOTA, testutil.RequestFundsAmount-3)
}
