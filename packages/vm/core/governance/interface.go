// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// in the blocklog core contract the VM keeps indices of blocks and requests in an optimized way
// for fast checking and timestamp access.
package governance

import (
	"github.com/iotaledger/wasp/packages/isc/coreutil"
)

var Contract = coreutil.NewContract(coreutil.CoreContractGovernance)

var (
	// state controller (entity that owns the state output via AliasAddress)
	FuncRotateStateController               = coreutil.Func(coreutil.CoreEPRotateStateController)
	FuncAddAllowedStateControllerAddress    = coreutil.Func("addAllowedStateControllerAddress")
	FuncRemoveAllowedStateControllerAddress = coreutil.Func("removeAllowedStateControllerAddress")
	ViewGetAllowedStateControllerAddresses  = coreutil.ViewFunc("getAllowedStateControllerAddresses")

	// chain owner (L1 entity that is the "owner of the chain")
	FuncClaimChainOwnership        = coreutil.Func("claimChainOwnership")
	FuncDelegateChainOwnership     = coreutil.Func("delegateChainOwnership")
	FuncSetPayoutAgentID           = coreutil.Func("setPayoutAgentID")
	FuncSetMinCommonAccountBalance = coreutil.Func("setMinCommonAccountBalance")
	ViewGetPayoutAgentID           = coreutil.ViewFunc("getPayoutAgentID")
	ViewGetMinCommonAccountBalance = coreutil.ViewFunc("getMinCommonAccountBalance")
	ViewGetChainOwner              = coreutil.ViewFunc("getChainOwner")

	// gas
	FuncSetFeePolicy = coreutil.Func("setFeePolicy")
	FuncSetGasLimits = coreutil.Func("setGasLimits")
	ViewGetFeePolicy = coreutil.ViewFunc("getFeePolicy")
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

	// public chain metadata
	FuncSetMetadata = coreutil.Func("setMetadata")
	ViewGetMetadata = coreutil.ViewFunc("getMetadata")
)

// state variables
const (
	// state controller
	VarAllowedStateControllerAddresses = "a" // covered in: TestGovernance1
	VarRotateToAddress                 = "r" // should never persist in the state

	VarPayoutAgentID                = "pa" // covered in: TestMetadata
	VarMinBaseTokensOnCommonAccount = "vs" // covered in: TestMetadata

	// chain owner
	VarChainOwnerID          = "o" // covered in: TestMetadata
	VarChainOwnerIDDelegated = "n" // covered in: TestMaintenanceMode

	// gas
	VarGasFeePolicyBytes = "g" // covered in: TestMetadata
	VarGasLimitsBytes    = "l" // covered in: TestMetadata

	// access nodes
	VarAccessNodes          = "an" // covered in: TestAccessNodes
	VarAccessNodeCandidates = "ac" // covered in: TestAccessNodes

	// maintenance
	VarMaintenanceStatus = "m" // covered in: TestMetadata

	// L2 metadata (provided by the webapi, located by the public url)
	VarMetadata = "md" // covered in: TestMetadata

	// L1 metadata (stored and provided in the tangle)
	VarPublicURL = "x" // covered in: TestL1Metadata

	// state pruning
	VarBlockKeepAmount = "b" // covered in: TestMetadata
)

// request parameters
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

	// public chain metadata (provided by the webapi, located by the public url)
	ParamMetadata = "md"

	// L1 metadata (stored and provided in the tangle)
	ParamPublicURL = "x"

	// state pruning
	ParamBlockKeepAmount = "b"

	// set payout AgentID
	ParamSetPayoutAgentID = "s"

	// set min SD
	ParamSetMinCommonAccountBalance = "ms"
)

// contract constants
const (
	// DefaultMinBaseTokensOnCommonAccount can't harvest the minimum
	DefaultMinBaseTokensOnCommonAccount = uint64(3000)

	BlockKeepAll           = -1
	DefaultBlockKeepAmount = 10_000
)
