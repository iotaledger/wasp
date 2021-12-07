// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// in the blocklog core contract the VM keeps indices of blocks and requests in an optimized way
// for fast checking and timestamp access.
package governance

import (
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
)

// constants
const (
	MinEventSize               = uint16(200)
	MinEventsPerRequest        = uint16(10)
	DefaultMaxEventsPerRequest = uint16(50)
	DefaultMaxEventSize        = uint16(2000)    // 2Kb
	DefaultMaxBlobSize         = uint32(1000000) // 1Mb
)

var Contract = coreutil.NewContract(coreutil.CoreContractGovernance, "Governance contract")

var (
	// state controller (entity that owns the state output via AliasAddress)
	FuncRotateStateController               = coreutil.Func(coreutil.CoreEPRotateStateController)
	FuncAddAllowedStateControllerAddress    = coreutil.Func("addAllowedStateControllerAddress")
	FuncRemoveAllowedStateControllerAddress = coreutil.Func("removeAllowedStateControllerAddress")
	FuncGetAllowedStateControllerAddresses  = coreutil.ViewFunc("getAllowedStateControllerAddresses")

	// chain owner (L1 entity that is the "owner of the chain")
	FuncClaimChainOwnership    = coreutil.Func("claimChainOwnership")
	FuncDelegateChainOwnership = coreutil.Func("delegateChainOwnership")
	FuncGetChainOwner          = coreutil.ViewFunc("getChainOwner")

	// fees
	FuncSetContractFee = coreutil.Func("setContractFee")
	FuncGetFeeInfo     = coreutil.ViewFunc("getFeeInfo")

	// chain info
	FuncSetChainInfo   = coreutil.Func("setChainInfo")
	FuncGetChainInfo   = coreutil.ViewFunc("getChainInfo")
	FuncGetMaxBlobSize = coreutil.ViewFunc("getMaxBlobSize")

	// access nodes
	FuncGetChainNodes     = coreutil.ViewFunc("getChainNodes")
	FuncAddCandidateNode  = coreutil.Func("addCandidateNode")
	FuncRevokeAccessNode  = coreutil.Func("revokeAccessNode")
	FuncChangeAccessNodes = coreutil.Func("changeAccessNodes")
)

// state variables
const (
	// state controller
	StateVarAllowedStateControllerAddresses = "a"
	StateVarRotateToAddress                 = "r"

	// chain owner
	VarChainOwnerID          = "o"
	VarChainOwnerIDDelegated = "n"
	VarDefaultOwnerFee       = "do"
	VarOwnerFee              = "of"

	// fees
	VarDefaultValidatorFee  = "dv"
	VarValidatorFee         = "vf"
	VarFeeAssetID           = "f"
	VarContractFeesRegistry = "fr"

	// chain info
	VarChainID         = "c"
	VarDescription     = "d"
	VarMaxBlobSize     = "mb"
	VarMaxEventSize    = "me"
	VarMaxEventsPerReq = "mr"

	// access nodes
	VarAccessNodes          = "an"
	VarAccessNodeCandidates = "ac"
	VarValidatorNodes       = "vn"
)

// params
const (
	// state controller
	ParamStateControllerAddress          = coreutil.ParamStateControllerAddress
	ParamAllowedStateControllerAddresses = kv.Key('a' + iota)

	// chain owner
	ParamChainOwner
	ParamOwnerFee

	// fees
	ParamFeeColor
	ParamValidatorFee
	ParamHname

	// chain info
	ParamChainID
	ParamDescription
	ParamMaxBlobSize
	ParamMaxEventSize
	ParamMaxEventsPerRequest

	ParamGetChainNodesAccessNodeCandidates
	ParamGetChainNodesAccessNodes

	// access nodes: addCandidateNode
	ParamAccessNodeInfoForCommittee
	ParamAccessNodeInfoPubKey
	ParamAccessNodeInfoCertificate
	ParamAccessNodeInfoAccessAPI

	// access nodes: changeAccessNodes
	ParamChangeAccessNodesActions
)

func init() {}
