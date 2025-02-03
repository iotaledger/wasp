// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// in the blocklog core contract the VM keeps indices of blocks and requests in an optimized way
// for fast checking and timestamp access.
package governance

import (
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

var Contract = coreutil.NewContract(coreutil.CoreContractGovernance)

var (
	// state controller (entity that owns the state output via AliasAddress)
	FuncRotateStateController = coreutil.NewEP1(Contract, coreutil.CoreEPRotateStateController,
		coreutil.Field[*cryptolib.Address](),
	)
	FuncAddAllowedStateControllerAddress = coreutil.NewEP1(Contract, "addAllowedStateControllerAddress",
		coreutil.Field[*cryptolib.Address](),
	)
	FuncRemoveAllowedStateControllerAddress = coreutil.NewEP1(Contract, "removeAllowedStateControllerAddress",
		coreutil.Field[*cryptolib.Address](),
	)
	ViewGetAllowedStateControllerAddresses = coreutil.NewViewEP01(Contract, "getAllowedStateControllerAddresses",
		coreutil.Field[[]*cryptolib.Address](),
	)

	// chain owner (L1 entity that is the "owner of the chain")
	FuncClaimChainOwnership    = coreutil.NewEP0(Contract, "claimChainOwnership")
	FuncDelegateChainOwnership = coreutil.NewEP1(Contract, "delegateChainOwnership",
		coreutil.Field[isc.AgentID](),
	)
	FuncSetPayoutAgentID = coreutil.NewEP1(Contract, "setPayoutAgentID",
		coreutil.Field[isc.AgentID](),
	)
	FuncSetGasCoinTargetValue = coreutil.NewEP1(Contract, "setGasCoinTargetValue",
		coreutil.Field[coin.Value](),
	)
	ViewGetPayoutAgentID = coreutil.NewViewEP01(Contract, "getPayoutAgentID",
		coreutil.Field[isc.AgentID](),
	)
	ViewGetGasCoinTargetValue = coreutil.NewViewEP01(Contract, "getGasCoinTargetValue",
		coreutil.Field[coin.Value](),
	)
	ViewGetChainOwner = coreutil.NewViewEP01(Contract, "getChainOwner",
		coreutil.Field[isc.AgentID](),
	)

	// gas
	FuncSetFeePolicy = coreutil.NewEP1(Contract, "setFeePolicy",
		coreutil.Field[*gas.FeePolicy](),
	)
	FuncSetGasLimits = coreutil.NewEP1(Contract, "setGasLimits",
		coreutil.Field[*gas.Limits](),
	)
	ViewGetFeePolicy = coreutil.NewViewEP01(Contract, "getFeePolicy",
		coreutil.Field[*gas.FeePolicy](),
	)
	ViewGetGasLimits = coreutil.NewViewEP01(Contract, "getGasLimits",
		coreutil.Field[*gas.Limits](),
	)

	// evm fees
	FuncSetEVMGasRatio = coreutil.NewEP1(Contract, "setEVMGasRatio",
		coreutil.Field[util.Ratio32](),
	)
	ViewGetEVMGasRatio = coreutil.NewViewEP01(Contract, "getEVMGasRatio",
		coreutil.Field[util.Ratio32](),
	)

	// chain info
	ViewGetChainInfo = coreutil.NewViewEP01(Contract, "getChainInfo",
		coreutil.Field[*isc.ChainInfo](),
	)

	// access nodes
	FuncAddCandidateNode = coreutil.NewEP4(Contract, "addCandidateNode",
		coreutil.Field[*cryptolib.PublicKey](), // NodePubKey
		coreutil.Field[[]byte](),               // Certificate
		coreutil.Field[string](),               // AccessAPI
		coreutil.Field[bool](),                 // ForCommittee
	)
	FuncRevokeAccessNode = coreutil.NewEP2(Contract, "revokeAccessNode",
		coreutil.Field[*cryptolib.PublicKey](), // NodePubKey
		coreutil.Field[[]byte](),               // Certificate
	)
	FuncChangeAccessNodes = coreutil.NewEP1(Contract, "changeAccessNodes",
		coreutil.Field[ChangeAccessNodeActions](),
	)
	ViewGetChainNodes = coreutil.NewViewEP02(Contract, "getChainNodes",
		coreutil.Field[[]*AccessNodeInfo](),
		coreutil.Field[[]*cryptolib.PublicKey](),
	)

	// maintenance
	FuncStartMaintenance     = coreutil.NewEP0(Contract, "startMaintenance")
	FuncStopMaintenance      = coreutil.NewEP0(Contract, "stopMaintenance")
	ViewGetMaintenanceStatus = coreutil.NewViewEP01(Contract, "getMaintenanceStatus",
		coreutil.Field[bool](),
	)

	// public chain metadata
	FuncSetMetadata = coreutil.NewEP2(Contract, "setMetadata",
		coreutil.FieldOptional[string](),
		coreutil.FieldOptional[*isc.PublicChainMetadata](),
	)
	ViewGetMetadata = coreutil.NewViewEP02(Contract, "getMetadata",
		coreutil.Field[string](),
		coreutil.Field[*isc.PublicChainMetadata](),
	)
)

// state variables
const (
	// state controller
	// varAllowedStateControllerAddresses :: map[Address]bool
	varAllowedStateControllerAddresses = "a" // covered in: TestGovernance1
	// varRotateToAddress :: Address (should never persist in the state)
	varRotateToAddress = "r"

	// VarPayoutAgentID :: AgentID
	VarPayoutAgentID = "pa" // covered in: TestMetadata
	// varGasCoinTargetValue :: uint64
	varGasCoinTargetValue = "vs" // covered in: TestMetadata

	// chain owner
	// VarChainOwnerID :: AgentID
	VarChainOwnerID = "o" // covered in: TestMetadata
	// VarChainOwnerIDDelegated :: AgentID
	VarChainOwnerIDDelegated = "n" // covered in: TestMaintenanceMode

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
	BlockKeepAll           = -1
	DefaultBlockKeepAmount = 10_000
)
