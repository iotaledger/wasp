// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// in the blocklog core contract the VM keeps indices of blocks and requests in an optimized way
// for fast checking and timestamp access.
package governance

import (
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

var Contract = coreutil.NewContract(coreutil.CoreContractGovernance)

var (
	// state controller (entity that owns the state output via AliasAddress)
	FuncRotateStateController = coreutil.NewEP1(Contract, coreutil.CoreEPRotateStateController,
		coreutil.FieldWithCodec(codec.Address),
	)
	FuncAddAllowedStateControllerAddress = coreutil.NewEP1(Contract, "addAllowedStateControllerAddress",
		coreutil.FieldWithCodec(codec.Address),
	)
	FuncRemoveAllowedStateControllerAddress = coreutil.NewEP1(Contract, "removeAllowedStateControllerAddress",
		coreutil.FieldWithCodec(codec.Address),
	)
	ViewGetAllowedStateControllerAddresses = coreutil.NewViewEP01(Contract, "getAllowedStateControllerAddresses",
		coreutil.FieldArrayWithCodec(codec.Address),
	)

	// chain owner (L1 entity that is the "owner of the chain")
	FuncClaimChainOwnership    = coreutil.NewEP0(Contract, "claimChainOwnership")
	FuncDelegateChainOwnership = coreutil.NewEP1(Contract, "delegateChainOwnership",
		coreutil.FieldWithCodec(codec.AgentID),
	)
	FuncSetPayoutAgentID = coreutil.NewEP1(Contract, "setPayoutAgentID",
		coreutil.FieldWithCodec(codec.AgentID),
	)
	FuncSetMinCommonAccountBalance = coreutil.NewEP1(Contract, "setMinCommonAccountBalance",
		coreutil.FieldWithCodec(codec.CoinValue),
	)
	ViewGetPayoutAgentID = coreutil.NewViewEP01(Contract, "getPayoutAgentID",
		coreutil.FieldWithCodec(codec.AgentID),
	)
	ViewGetMinCommonAccountBalance = coreutil.NewViewEP01(Contract, "getMinCommonAccountBalance",
		coreutil.FieldWithCodec(codec.CoinValue),
	)
	ViewGetChainOwner = coreutil.NewViewEP01(Contract, "getChainOwner",
		coreutil.FieldWithCodec(codec.AgentID),
	)

	// gas
	FuncSetFeePolicy = coreutil.NewEP1(Contract, "setFeePolicy",
		coreutil.FieldWithCodec(codec.NewCodecEx(gas.FeePolicyFromBytes)),
	)
	FuncSetGasLimits = coreutil.NewEP1(Contract, "setGasLimits",
		coreutil.FieldWithCodec(codec.NewCodecEx(gas.LimitsFromBytes)),
	)
	ViewGetFeePolicy = coreutil.NewViewEP01(Contract, "getFeePolicy",
		coreutil.FieldWithCodec(codec.NewCodecEx(gas.FeePolicyFromBytes)),
	)
	ViewGetGasLimits = coreutil.NewViewEP01(Contract, "getGasLimits",
		coreutil.FieldWithCodec(codec.NewCodecEx(gas.LimitsFromBytes)),
	)

	// evm fees
	FuncSetEVMGasRatio = coreutil.NewEP1(Contract, "setEVMGasRatio",
		coreutil.FieldWithCodec(codec.NewCodecEx(util.Ratio32FromBytes)),
	)
	ViewGetEVMGasRatio = coreutil.NewViewEP01(Contract, "getEVMGasRatio",
		coreutil.FieldWithCodec(codec.NewCodecEx(util.Ratio32FromBytes)),
	)

	// chain info
	ViewGetChainInfo = coreutil.NewViewEP01(Contract, "getChainInfo",
		coreutil.FieldWithCodec(codec.NewCodecEx(isc.ChainInfoFromBytes)),
	)

	// access nodes
	FuncAddCandidateNode = coreutil.NewEP4(Contract, "addCandidateNode",
		coreutil.FieldWithCodec(codec.NewCodecFromBCS[*cryptolib.PublicKey]()), // NodePubKey
		coreutil.FieldWithCodec(codec.Bytes),                                   // Certificate
		coreutil.FieldWithCodec(codec.String),                                  // AccessAPI
		coreutil.FieldWithCodec(codec.Bool),                                    // ForCommittee
	)
	FuncRevokeAccessNode = coreutil.NewEP2(Contract, "revokeAccessNode",
		coreutil.FieldWithCodec(codec.NewCodecFromBCS[*cryptolib.PublicKey]()), // NodePubKey
		coreutil.FieldWithCodec(codec.Bytes),                                   // Certificate
	)
	FuncChangeAccessNodes = coreutil.NewEP1(Contract, "changeAccessNodes",
		coreutil.FieldArrayWithCodec(codec.NewTupleCodec[*cryptolib.PublicKey, *ChangeAccessNodeAction]()),
	)
	ViewGetChainNodes = coreutil.NewViewEP02(Contract, "getChainNodes",
		coreutil.FieldArrayWithCodec(codec.NewCodecFromBCS[*AccessNodeInfo]()),
		coreutil.FieldArrayWithCodec(codec.NewCodecEx(cryptolib.PublicKeyFromBytes)),
	)

	// maintenance
	FuncStartMaintenance     = coreutil.NewEP0(Contract, "startMaintenance")
	FuncStopMaintenance      = coreutil.NewEP0(Contract, "stopMaintenance")
	ViewGetMaintenanceStatus = coreutil.NewViewEP01(Contract, "getMaintenanceStatus",
		coreutil.FieldWithCodec(codec.Bool),
	)

	// public chain metadata
	FuncSetMetadata = coreutil.NewEP2(Contract, "setMetadata",
		coreutil.FieldWithCodecOptional(codec.String),
		coreutil.FieldWithCodecOptional(codec.NewCodecEx(isc.PublicChainMetadataFromBytes)),
	)
	ViewGetMetadata = coreutil.NewViewEP02(Contract, "getMetadata",
		coreutil.FieldWithCodec(codec.String),
		coreutil.FieldWithCodec(codec.NewCodecEx(isc.PublicChainMetadataFromBytes)),
	)
)

// state variables
const (
	// state controller
	// varAllowedStateControllerAddresses :: map[Address]bool
	varAllowedStateControllerAddresses = "a" // covered in: TestGovernance1
	// varRotateToAddress :: Address (should never persist in the state)
	varRotateToAddress = "r"

	// varPayoutAgentID :: AgentID
	varPayoutAgentID = "pa" // covered in: TestMetadata
	// varMinBaseTokensOnCommonAccount :: uint64
	varMinBaseTokensOnCommonAccount = "vs" // covered in: TestMetadata

	// chain owner
	// varChainOwnerID :: AgentID
	varChainOwnerID = "o" // covered in: TestMetadata
	// varChainOwnerIDDelegated :: AgentID
	varChainOwnerIDDelegated = "n" // covered in: TestMaintenanceMode

	// gas
	// varGasFeePolicyBytes :: gas.FeePolicy
	varGasFeePolicyBytes = "g" // covered in: TestMetadata
	// varGasLimitsBytes :: gas.Limits
	varGasLimitsBytes = "l" // covered in: TestMetadata

	// access nodes
	// varAccessNodes :: map[PublicKey]bool
	varAccessNodes = "an" // covered in: TestAccessNodes
	// varAccessNodes :: map[PublicKey]AccessNodeData
	varAccessNodeCandidates = "ac" // covered in: TestAccessNodes

	// maintenance
	// varMaintenanceStatus :: bool
	varMaintenanceStatus = "m" // covered in: TestMetadata

	// L2 metadata (provided by the webapi, located by the public url)
	// varMetadata :: isc.PublicChainMetadata
	varMetadata = "md" // covered in: TestMetadata

	// L1 metadata (stored and provided in the tangle)
	// varPublicURL :: string
	varPublicURL = "x" // covered in: TestL1Metadata

	// state pruning
	// varBlockKeepAmount :: int32
	varBlockKeepAmount = "b" // covered in: TestMetadata
)

// contract constants
const (
	// DefaultMinBaseTokensOnCommonAccount can't harvest the minimum
	DefaultMinBaseTokensOnCommonAccount = 3000

	BlockKeepAll           = -1
	DefaultBlockKeepAmount = 10_000
)
