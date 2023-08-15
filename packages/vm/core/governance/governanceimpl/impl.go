// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package governanceimpl

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

var Processor = governance.Contract.Processor(nil,
	// state controller
	governance.FuncAddAllowedStateControllerAddress.WithHandler(addAllowedStateControllerAddress),
	governance.FuncRemoveAllowedStateControllerAddress.WithHandler(removeAllowedStateControllerAddress),
	governance.FuncRotateStateController.WithHandler(rotateStateController),
	governance.ViewGetAllowedStateControllerAddresses.WithHandler(getAllowedStateControllerAddresses),

	// chain owner
	governance.FuncClaimChainOwnership.WithHandler(claimChainOwnership),
	governance.FuncDelegateChainOwnership.WithHandler(delegateChainOwnership),
	governance.FuncSetPayoutAgentID.WithHandler(setPayoutAgentID),
	governance.ViewGetPayoutAgentID.WithHandler(getPayoutAgentID),
	governance.FuncSetMinCommonAccountBalance.WithHandler(setMinCommonAccountBalance),
	governance.ViewGetMinCommonAccountBalance.WithHandler(getMinCommonAccountBalance),
	governance.ViewGetChainOwner.WithHandler(getChainOwner),

	// fees
	governance.FuncSetFeePolicy.WithHandler(setFeePolicy),
	governance.ViewGetFeePolicy.WithHandler(getFeePolicy),
	governance.FuncSetEVMGasRatio.WithHandler(setEVMGasRatio),
	governance.ViewGetEVMGasRatio.WithHandler(getEVMGasRatio),
	governance.FuncSetGasLimits.WithHandler(setGasLimits),
	governance.ViewGetGasLimits.WithHandler(getGasLimits),

	// chain info
	governance.ViewGetChainInfo.WithHandler(getChainInfo),

	// access nodes
	governance.FuncAddCandidateNode.WithHandler(addCandidateNode),
	governance.FuncChangeAccessNodes.WithHandler(changeAccessNodes),
	governance.FuncRevokeAccessNode.WithHandler(revokeAccessNode),
	governance.ViewGetChainNodes.WithHandler(getChainNodes),

	// maintenance
	governance.FuncStartMaintenance.WithHandler(startMaintenance),
	governance.FuncStopMaintenance.WithHandler(stopMaintenance),
	governance.ViewGetMaintenanceStatus.WithHandler(getMaintenanceStatus),

	// L1 metadata
	governance.FuncSetMetadata.WithHandler(setMetadata),
	governance.ViewGetMetadata.WithHandler(getMetadata),
)

func SetInitialState(state kv.KVStore, chainOwner isc.AgentID, blockKeepAmount int32) {
	state.Set(governance.VarChainOwnerID, chainOwner.Bytes())
	state.Set(governance.VarGasFeePolicyBytes, gas.DefaultFeePolicy().Bytes())
	state.Set(governance.VarGasLimitsBytes, gas.LimitsDefault.Bytes())
	state.Set(governance.VarMaintenanceStatus, codec.Encode(false))
	state.Set(governance.VarBlockKeepAmount, codec.EncodeInt32(blockKeepAmount))
	state.Set(governance.VarMinBaseTokensOnCommonAccount, codec.EncodeUint64(governance.DefaultMinBaseTokensOnCommonAccount))
	state.Set(governance.VarPayoutAgentID, chainOwner.Bytes())
}
