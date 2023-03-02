// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// in the blocklog core contract the VM keeps indices of blocks and requests in an optimized way
// for fast checking and timestamp access.
package governance

import (
	"github.com/iotaledger/wasp/packages/isc/coreutil"
)

// constants
const (
	MinEventSize               = uint16(200)
	MinEventsPerRequest        = uint16(10)
	DefaultMaxEventsPerRequest = uint16(50)
	DefaultMaxEventSize        = uint16(2000)      // 2Kb
	DefaultMaxBlobSize         = uint32(2_000_000) // 2Mb
)

var Contract = coreutil.NewContract(coreutil.CoreContractGovernance, "Governance contract")

var (
	// state controller (entity that owns the state output via AliasAddress)
	FuncRotateStateController               = coreutil.Func(coreutil.CoreEPRotateStateController)
	FuncAddAllowedStateControllerAddress    = coreutil.Func("addAllowedStateControllerAddress")
	FuncRemoveAllowedStateControllerAddress = coreutil.Func("removeAllowedStateControllerAddress")
	ViewGetAllowedStateControllerAddresses  = coreutil.ViewFunc("getAllowedStateControllerAddresses")

	// chain owner (L1 entity that is the "owner of the chain")
	FuncClaimChainOwnership    = coreutil.Func("claimChainOwnership")
	FuncDelegateChainOwnership = coreutil.Func("delegateChainOwnership")
	ViewGetChainOwner          = coreutil.ViewFunc("getChainOwner")

	// fees
	FuncSetFeePolicy = coreutil.Func("setFeePolicy")
	ViewGetFeePolicy = coreutil.ViewFunc("getFeePolicy")
	// evm fees
	FuncSetEVMGasRatio = coreutil.Func("setGasRatio")
	ViewGetEVMGasRatio = coreutil.ViewFunc("getGasRatio")

	// chain info
	FuncSetChainInfo   = coreutil.Func("setChainInfo")
	ViewGetChainInfo   = coreutil.ViewFunc("getChainInfo")
	ViewGetMaxBlobSize = coreutil.ViewFunc("getMaxBlobSize")

	// access nodes
	FuncAddCandidateNode  = coreutil.Func("addCandidateNode")
	FuncRevokeAccessNode  = coreutil.Func("revokeAccessNode")
	FuncChangeAccessNodes = coreutil.Func("changeAccessNodes")
	ViewGetChainNodes     = coreutil.ViewFunc("getChainNodes")

	// maintenance
	FuncStartMaintenance     = coreutil.Func("startMaintenance")
	FuncStopMaintenance      = coreutil.Func("stopMaintenance")
	ViewGetMaintenanceStatus = coreutil.ViewFunc("getMaintenanceStatus")

	// L1 metadata
	FuncSetCustomMetadata = coreutil.Func("setCustomMetadata")
	ViewGetCustomMetadata = coreutil.ViewFunc("getCustomMetadata")
)

// state variables
const (
	// state controller
	StateVarAllowedStateControllerAddresses = "a"
	StateVarRotateToAddress                 = "r"

	// chain owner
	VarChainOwnerID          = "o"
	VarChainOwnerIDDelegated = "n"

	// fees
	VarGasFeePolicyBytes = "g"

	// chain info
	VarChainID         = "c"
	VarDescription     = "d"
	VarMaxBlobSize     = "mb"
	VarMaxEventSize    = "me"
	VarMaxEventsPerReq = "mr"

	// access nodes
	VarAccessNodes          = "an"
	VarAccessNodeCandidates = "ac"

	// maintenance
	VarMaintenanceStatus = "m"

	// L1 metadata
	VarCustomMetadata = "x"
)

// params
const (
	// state controller
	ParamStateControllerAddress          = coreutil.ParamStateControllerAddress
	ParamAllowedStateControllerAddresses = "a"

	// chain owner
	ParamChainOwner = "o"

	// fees
	ParamFeePolicyBytes = "g"
	ParamEVMGasRatio    = "e"

	// chain info
	ParamChainID                   = "c"
	ParamDescription               = "d"
	ParamMaxBlobSizeUint32         = "mb"
	ParamMaxEventSizeUint16        = "me"
	ParamMaxEventsPerRequestUint16 = "mr"

	ParamGetChainNodesAccessNodeCandidates = "an"
	ParamGetChainNodesAccessNodes          = "ac"

	// access nodes: addCandidateNode
	ParamAccessNodeInfoForCommittee = "i"
	ParamAccessNodeInfoPubKey       = "ip"
	ParamAccessNodeInfoCertificate  = "ic"
	ParamAccessNodeInfoAccessAPI    = "ia"

	// access nodes: changeAccessNodes
	ParamChangeAccessNodesActions = "n"

	// L1 metadata
	ParamCustomMetadata = "x"
)
