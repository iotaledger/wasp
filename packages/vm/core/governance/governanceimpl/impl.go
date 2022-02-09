// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package governanceimpl

import (
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

var Processor = governance.Contract.Processor(initialize,
	// state controller
	governance.FuncRotateStateController.WithHandler(rotateStateController),
	governance.FuncAddAllowedStateControllerAddress.WithHandler(addAllowedStateControllerAddress),
	governance.FuncRemoveAllowedStateControllerAddress.WithHandler(removeAllowedStateControllerAddress),
	governance.FuncGetAllowedStateControllerAddresses.WithHandler(getAllowedStateControllerAddresses),

	// chain owner
	governance.FuncClaimChainOwnership.WithHandler(claimChainOwnership),
	governance.FuncDelegateChainOwnership.WithHandler(delegateChainOwnership),
	governance.FuncGetChainOwner.WithHandler(getChainOwner),

	// fees
	governance.FuncSetFeePolicy.WithHandler(setFeePolicy),
	governance.FuncGetFeePolicy.WithHandler(getFeePolicy),

	// chain info
	governance.FuncGetChainInfo.WithHandler(getChainInfo),
	governance.FuncSetChainInfo.WithHandler(setChainInfo),
	governance.FuncGetMaxBlobSize.WithHandler(getMaxBlobSize),

	// access nodes.
	governance.FuncGetChainNodes.WithHandler(getChainNodesFuncHandler),
	governance.FuncAddCandidateNode.WithHandler(addCandidateNodeFuncHandler),
	governance.FuncRevokeAccessNode.WithHandler(revokeAccessNodeFuncHandler),
	governance.FuncChangeAccessNodes.WithHandler(changeAccessNodesFuncHandler),
)

func initialize(ctx iscp.Sandbox) dict.Dict {
	ctx.Log().Debugf("governance.initialize.begin")
	state := ctx.State()

	// retrieving init parameters
	// -- chain ID

	chainID := ctx.ParamDecoder().MustGetChainID(governance.ParamChainID)
	chainDescription := ctx.ParamDecoder().MustGetString(governance.ParamDescription, "N/A")
	feePolicyBytes := ctx.ParamDecoder().MustGetBytes(governance.ParamFeePolicyBytes, gas.DefaultGasFeePolicy().Bytes())

	state.Set(governance.VarChainID, codec.EncodeChainID(chainID))
	state.Set(governance.VarChainOwnerID, ctx.ParamDecoder().MustGetAgentID(governance.ParamChainOwner).Bytes())
	state.Set(governance.VarDescription, codec.EncodeString(chainDescription))

	state.Set(governance.VarMaxBlobSize, codec.Encode(governance.DefaultMaxBlobSize))
	state.Set(governance.VarMaxEventSize, codec.Encode(governance.DefaultMaxEventSize))
	state.Set(governance.VarMaxEventsPerReq, codec.Encode(governance.DefaultMaxEventsPerRequest))

	state.Set(governance.VarGasFeePolicyBytes, feePolicyBytes)

	return nil
}
