// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmtest

import (
	"testing"

	"github.com/iotaledger/wasp/contracts/native/evm"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/assert"
	"github.com/iotaledger/wasp/packages/iscp/colored"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/stretchr/testify/require"
)

// TODO this SC could adjust gasPerIota based on some conditions

var (
	evmChainMgmtContract = coreutil.NewContract("EVMChainManagement", "EVM chain management")

	mgmtFuncClaimOwnership  = coreutil.Func("claimOwnership")
	mgmtFuncWithdrawGasFees = coreutil.Func("withdrawGasFees")
)

func TestRequestGasFees(t *testing.T) {
	withEVMFlavors(t, func(t *testing.T, evmFlavor *coreutil.ContractInfo) {
		evmChainMgmtProcessor := evmChainMgmtContract.Processor(nil,
			mgmtFuncClaimOwnership.WithHandler(func(ctx iscp.Sandbox) (dict.Dict, error) {
				a := assert.NewAssert(ctx.Log())
				_, err := ctx.Call(evmFlavor.Hname(), evm.FuncClaimOwnership.Hname(), nil, nil)
				a.RequireNoError(err)
				return nil, nil
			}),
			mgmtFuncWithdrawGasFees.WithHandler(func(ctx iscp.Sandbox) (dict.Dict, error) {
				a := assert.NewAssert(ctx.Log())
				_, err := ctx.Call(evmFlavor.Hname(), evm.FuncWithdrawGasFees.Hname(), nil, nil)
				a.RequireNoError(err)
				return nil, nil
			}),
		)

		evmChain := initEVMChain(t, evmFlavor, evmChainMgmtProcessor)
		soloChain := evmChain.soloChain

		err := soloChain.DeployContract(nil, evmChainMgmtContract.Name, evmChainMgmtContract.ProgramHash)
		require.NoError(t, err)

		// deploy solidity `storage` contract (just to produce some fees to be claimed)
		evmChain.deployStorageContract(evmChain.faucetKey, 42)

		// change owner to evnchainmanagement SC
		managerAgentID := iscp.NewAgentID(soloChain.ChainID.AsAddress(), iscp.Hn(evmChainMgmtContract.Name))
		_, err = soloChain.PostRequestSync(
			solo.NewCallParams(evmFlavor.Name, evm.FuncSetNextOwner.Name, evm.FieldNextEVMOwner, managerAgentID).
				WithIotas(1),
			soloChain.OriginatorKeyPair,
		)
		require.NoError(t, err)

		// claim ownership
		_, err = soloChain.PostRequestSync(
			solo.NewCallParams(evmChainMgmtContract.Name, mgmtFuncClaimOwnership.Name).WithIotas(1),
			soloChain.OriginatorKeyPair,
		)
		require.NoError(t, err)

		// call requestGasFees manually, so that the manager SC request funds from the evm chain, check funds are received by the manager SC
		balance0 := soloChain.GetAccountBalance(managerAgentID).Get(colored.IOTA)

		_, err = soloChain.PostRequestSync(
			solo.NewCallParams(evmChainMgmtContract.Name, mgmtFuncWithdrawGasFees.Name).WithIotas(1),
			soloChain.OriginatorKeyPair,
		)
		require.NoError(t, err)
		balance1 := soloChain.GetAccountBalance(managerAgentID).Get(colored.IOTA)

		require.Greater(t, balance1, balance0)
	})
}
