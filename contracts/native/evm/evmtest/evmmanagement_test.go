// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmtest

import (
	"testing"

	"github.com/iotaledger/wasp/contracts/native/evm"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/stretchr/testify/require"
)

// TODO this SC could adjust gasPerIota based on some conditions

var (
	evmChainMgmtContract = coreutil.NewContract("EVMChainManagement", "EVM chain management")

	mgmtFuncClaimOwnership = coreutil.Func("claimOwnership")
)

func TestRequestGasFees(t *testing.T) {
	withEVMFlavors(t, func(t *testing.T, evmFlavor *coreutil.ContractInfo) {
		evmChainMgmtProcessor := evmChainMgmtContract.Processor(nil,
			mgmtFuncClaimOwnership.WithHandler(func(ctx iscp.Sandbox) dict.Dict {
				ctx.Call(evmFlavor.Hname(), evm.FuncClaimOwnership.Hname(), nil, nil)
				return nil
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
				AddAssetsIotas(1),
			&soloChain.OriginatorPrivateKey,
		)
		require.NoError(t, err)

		// claim ownership
		_, err = soloChain.PostRequestSync(
			solo.NewCallParams(evmChainMgmtContract.Name, mgmtFuncClaimOwnership.Name).AddAssetsIotas(1),
			&soloChain.OriginatorPrivateKey,
		)
		require.NoError(t, err)
	})
}
