// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package governanceimpl

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

var Processor = governance.Contract.Processor(initialize,
	// state controller
	governance.FuncAddAllowedStateControllerAddress.WithHandler(addAllowedStateControllerAddress),
	governance.FuncRemoveAllowedStateControllerAddress.WithHandler(removeAllowedStateControllerAddress),
	governance.FuncRotateStateController.WithHandler(rotateStateController),
	governance.ViewGetAllowedStateControllerAddresses.WithHandler(getAllowedStateControllerAddresses),

	// chain owner
	governance.FuncClaimChainOwnership.WithHandler(claimChainOwnership),
	governance.FuncDelegateChainOwnership.WithHandler(delegateChainOwnership),
	governance.ViewGetChainOwner.WithHandler(getChainOwner),

	// fees
	governance.FuncSetFeePolicy.WithHandler(setFeePolicy),
	governance.ViewGetFeePolicy.WithHandler(getFeePolicy),

	// chain info
	governance.FuncSetChainInfo.WithHandler(setChainInfo),
	governance.ViewGetChainInfo.WithHandler(getChainInfo),
	governance.ViewGetMaxBlobSize.WithHandler(getMaxBlobSize),

	// access nodes
	governance.FuncAddCandidateNode.WithHandler(addCandidateNode),
	governance.FuncChangeAccessNodes.WithHandler(changeAccessNodes),
	governance.FuncRevokeAccessNode.WithHandler(revokeAccessNode),
	governance.ViewGetChainNodes.WithHandler(getChainNodes),

	// maintenance
	governance.FuncStartMaintenance.WithHandler(setMaintenanceOn),
	governance.FuncStopMaintenance.WithHandler(setMaintenanceOff),
	governance.ViewGetMaintenanceStatus.WithHandler(getMaintenanceStatus),
)

func initialize(ctx isc.Sandbox) dict.Dict {
	ctx.Log().Debugf("governance.initialize.begin")
	state := ctx.State()

	// retrieving init parameters
	// -- chain ID

	chainID := ctx.Params().MustGetChainID(governance.ParamChainID)
	chainDescription := ctx.Params().MustGetString(governance.ParamDescription, "N/A")
	feePolicyBytes := ctx.Params().MustGetBytes(governance.ParamFeePolicyBytes, gas.DefaultGasFeePolicy().Bytes())

	state.Set(governance.VarChainID, codec.EncodeChainID(chainID))
	state.Set(governance.VarChainOwnerID, ctx.Params().MustGetAgentID(governance.ParamChainOwner).Bytes())
	state.Set(governance.VarDescription, codec.EncodeString(chainDescription))

	state.Set(governance.VarMaxBlobSize, codec.Encode(governance.DefaultMaxBlobSize))
	state.Set(governance.VarMaxEventSize, codec.Encode(governance.DefaultMaxEventSize))
	state.Set(governance.VarMaxEventsPerReq, codec.Encode(governance.DefaultMaxEventsPerRequest))

	state.Set(governance.VarGasFeePolicyBytes, feePolicyBytes)

	state.Set(governance.VarMaintenanceStatus, codec.Encode(false))

	// storing hname as a terminal value of the contract's state root.
	// This way we will be able to retrieve commitment to the contract's state
	ctx.State().Set("", ctx.Contract().Bytes())

	return nil
}
