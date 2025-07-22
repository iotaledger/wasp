// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package governanceimpl

import (
	"github.com/iotaledger/wasp/v2/packages/vm/core/governance"
)

var Processor = governance.Contract.Processor(nil,
	// chain admin
	governance.FuncClaimChainAdmin.WithHandler(claimChainAdmin),
	governance.FuncDelegateChainAdmin.WithHandler(delegateChainAdmin),
	governance.FuncSetPayoutAgentID.WithHandler(setPayoutAgentID),
	governance.ViewGetPayoutAgentID.WithHandler(getPayoutAgentID),
	governance.FuncSetGasCoinTargetValue.WithHandler(setGasCoinTargetValue),
	governance.ViewGetGasCoinTargetValue.WithHandler(getGasCoinTargetValue),
	governance.ViewGetChainAdmin.WithHandler(getChainAdmin),

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
