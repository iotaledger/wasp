// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// in the blocklog core contract the VM keeps indices of blocks and requests in an optimized way
// for fast checking and timestamp access.
package governance

import (
	"github.com/iotaledger/wasp/packages/isc/coreutil"
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

	// gas
	FuncSetFeePolicy = coreutil.Func("setFeePolicy")
	ViewGetFeePolicy = coreutil.ViewFunc("getFeePolicy")
	FuncSetGasLimits = coreutil.Func("setGasLimits")
	ViewGetGasLimits = coreutil.ViewFunc("getGasLimits")

	// evm fees
	FuncSetEVMGasRatio = coreutil.Func("setEVMGasRatio")
	ViewGetEVMGasRatio = coreutil.ViewFunc("getEVMGasRatio")

	// chain info
	ViewGetChainInfo = coreutil.ViewFunc("getChainInfo")

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
	FuncSetMetadata = coreutil.Func("setMetadata")
	ViewGetMetadata = coreutil.ViewFunc("getMetadata")
)

// state variables
const (
	// state controller
	StateVarAllowedStateControllerAddresses = "a"
	StateVarRotateToAddress                 = "r"

	// chain owner
	VarChainOwnerID          = "o"
	VarChainOwnerIDDelegated = "n"

	// gas
	VarGasFeePolicyBytes = "g"
	VarGasLimitsBytes    = "l"

	// access nodes
	VarAccessNodes          = "an"
	VarAccessNodeCandidates = "ac"

	// maintenance
	VarMaintenanceStatus = "m"

	// L2 metadata (provided by the webapi, located by the public url)
	VarMetadataEVMJsonRPCURL   = "e"
	VarMetadataEVMWebSocketURL = "w"

	// L1 metadata (stored and provided in the tangle)
	VarPublicURL = "x"
)

// params
const (
	// state controller
	ParamStateControllerAddress          = coreutil.ParamStateControllerAddress
	ParamAllowedStateControllerAddresses = "a"

	// chain owner
	ParamChainOwner = "o"

	// gas
	ParamFeePolicyBytes = "g"
	ParamEVMGasRatio    = "e"
	ParamGasLimitsBytes = "l"

	// chain info
	ParamChainID = "c"

	ParamGetChainNodesAccessNodeCandidates = "an"
	ParamGetChainNodesAccessNodes          = "ac"

	// access nodes: addCandidateNode
	ParamAccessNodeInfoForCommittee = "i"
	ParamAccessNodeInfoPubKey       = "ip"
	ParamAccessNodeInfoCertificate  = "ic"
	ParamAccessNodeInfoAccessAPI    = "ia"

	// access nodes: changeAccessNodes
	ParamChangeAccessNodesActions = "n"

	// L2 metadata (provided by the webapi, located by the public url)
	ParamMetadataEVMJsonRPCURL   = "e"
	ParamMetadataEVMWebSocketURL = "w"

	// L1 metadata (stored and provided in the tangle)
	ParamPublicURL = "x"
)
