// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// in the blocklog core contract the VM keeps indices of blocks and requests in an optimized way
// for fast checking and timestamp access.
package governance

import (
	"errors"

	"github.com/samber/lo"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core/errors/coreerrors"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

var Contract = coreutil.NewContract(coreutil.CoreContractGovernance)

var (
	// state controller (entity that owns the state output via AliasAddress)
	FuncRotateStateController = coreutil.NewEP1(Contract, coreutil.CoreEPRotateStateController,
		coreutil.FieldWithCodec(ParamStateControllerAddress, codec.Address),
	)
	FuncAddAllowedStateControllerAddress = coreutil.NewEP1(Contract, "addAllowedStateControllerAddress",
		coreutil.FieldWithCodec(ParamStateControllerAddress, codec.Address),
	)
	FuncRemoveAllowedStateControllerAddress = coreutil.NewEP1(Contract, "removeAllowedStateControllerAddress",
		coreutil.FieldWithCodec(ParamStateControllerAddress, codec.Address),
	)
	ViewGetAllowedStateControllerAddresses = coreutil.NewViewEP01(Contract, "getAllowedStateControllerAddresses",
		OutputAddressList{},
	)

	// chain owner (L1 entity that is the "owner of the chain")
	FuncClaimChainOwnership    = coreutil.NewEP0(Contract, "claimChainOwnership")
	FuncDelegateChainOwnership = coreutil.NewEP1(Contract, "delegateChainOwnership",
		coreutil.FieldWithCodec(ParamChainOwner, codec.AgentID),
	)
	FuncSetPayoutAgentID = coreutil.NewEP1(Contract, "setPayoutAgentID",
		coreutil.FieldWithCodec(ParamSetPayoutAgentID, codec.AgentID),
	)
	FuncSetMinCommonAccountBalance = coreutil.NewEP1(Contract, "setMinCommonAccountBalance",
		coreutil.FieldWithCodec(ParamSetMinCommonAccountBalance, codec.Uint64),
	)
	ViewGetPayoutAgentID = coreutil.NewViewEP01(Contract, "getPayoutAgentID",
		coreutil.FieldWithCodec(ParamSetPayoutAgentID, codec.AgentID),
	)
	ViewGetMinCommonAccountBalance = coreutil.NewViewEP01(Contract, "getMinCommonAccountBalance",
		coreutil.FieldWithCodec(ParamSetMinCommonAccountBalance, codec.Uint64),
	)
	ViewGetChainOwner = coreutil.NewViewEP01(Contract, "getChainOwner",
		coreutil.FieldWithCodec(ParamChainOwner, codec.AgentID),
	)

	// gas
	FuncSetFeePolicy = coreutil.NewEP1(Contract, "setFeePolicy",
		coreutil.FieldWithCodec(ParamFeePolicyBytes, codec.NewCodecEx(gas.FeePolicyFromBytes)),
	)
	FuncSetGasLimits = coreutil.NewEP1(Contract, "setGasLimits",
		coreutil.FieldWithCodec(ParamGasLimitsBytes, codec.NewCodecEx(gas.LimitsFromBytes)),
	)
	ViewGetFeePolicy = coreutil.NewViewEP01(Contract, "getFeePolicy",
		coreutil.FieldWithCodec(ParamFeePolicyBytes, codec.NewCodecEx(gas.FeePolicyFromBytes)),
	)
	ViewGetGasLimits = coreutil.NewViewEP01(Contract, "getGasLimits",
		coreutil.FieldWithCodec(ParamGasLimitsBytes, codec.NewCodecEx(gas.LimitsFromBytes)),
	)

	// evm fees
	FuncSetEVMGasRatio = coreutil.NewEP1(Contract, "setEVMGasRatio",
		coreutil.FieldWithCodec(ParamEVMGasRatio, codec.NewCodecEx(util.Ratio32FromBytes)),
	)
	ViewGetEVMGasRatio = coreutil.NewViewEP01(Contract, "getEVMGasRatio",
		coreutil.FieldWithCodec(ParamEVMGasRatio, codec.NewCodecEx(util.Ratio32FromBytes)),
	)

	// chain info
	ViewGetChainInfo = coreutil.NewViewEP01(Contract, "getChainInfo",
		OutputChainInfo{},
	)

	// access nodes
	FuncAddCandidateNode = coreutil.NewEP1(Contract, "addCandidateNode",
		InputAddCandidateNode{},
	)
	FuncRevokeAccessNode = coreutil.NewEP1(Contract, "revokeAccessNode",
		InputRevokeAccessNode{},
	)
	FuncChangeAccessNodes = coreutil.NewEP1(Contract, "changeAccessNodes",
		InputChangeAccessNodes{},
	)
	ViewGetChainNodes = coreutil.NewViewEP01(Contract, "getChainNodes",
		OutputChainNodes{},
	)

	// maintenance
	FuncStartMaintenance     = coreutil.NewEP0(Contract, "startMaintenance")
	FuncStopMaintenance      = coreutil.NewEP0(Contract, "stopMaintenance")
	ViewGetMaintenanceStatus = coreutil.NewViewEP01(Contract, "getMaintenanceStatus",
		coreutil.FieldWithCodec(ParamMaintenanceStatus, codec.Bool),
	)

	// public chain metadata
	FuncSetMetadata = coreutil.NewEP2(Contract, "setMetadata",
		coreutil.FieldWithCodecOptional(ParamPublicURL, codec.String),
		coreutil.FieldWithCodecOptional(ParamMetadata, codec.NewCodecEx(isc.PublicChainMetadataFromBytes)),
	)
	ViewGetMetadata = coreutil.NewViewEP02(Contract, "getMetadata",
		coreutil.FieldWithCodec(ParamPublicURL, codec.String),
		coreutil.FieldWithCodec(ParamMetadata, codec.NewCodecEx(isc.PublicChainMetadataFromBytes)),
	)
)

// state variables
const (
	// state controller
	varAllowedStateControllerAddresses = "a" // covered in: TestGovernance1
	varRotateToAddress                 = "r" // should never persist in the state

	varPayoutAgentID                = "pa" // covered in: TestMetadata
	varMinBaseTokensOnCommonAccount = "vs" // covered in: TestMetadata

	// chain owner
	varChainOwnerID          = "o" // covered in: TestMetadata
	varChainOwnerIDDelegated = "n" // covered in: TestMaintenanceMode

	// gas
	varGasFeePolicyBytes = "g" // covered in: TestMetadata
	varGasLimitsBytes    = "l" // covered in: TestMetadata

	// access nodes
	varAccessNodes          = "an" // covered in: TestAccessNodes
	varAccessNodeCandidates = "ac" // covered in: TestAccessNodes

	// maintenance
	varMaintenanceStatus = "m" // covered in: TestMetadata

	// L2 metadata (provided by the webapi, located by the public url)
	varMetadata = "md" // covered in: TestMetadata

	// L1 metadata (stored and provided in the tangle)
	varPublicURL = "x" // covered in: TestL1Metadata

	// state pruning
	varBlockKeepAmount = "b" // covered in: TestMetadata
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

	ParamMaintenanceStatus = "m"
)

// contract constants
const (
	// DefaultMinBaseTokensOnCommonAccount can't harvest the minimum
	DefaultMinBaseTokensOnCommonAccount = uint64(3000)

	BlockKeepAll           = -1
	DefaultBlockKeepAmount = 10_000
)

type OutputAddressList struct{}

func (e OutputAddressList) Encode(addrs []*cryptolib.Address) dict.Dict {
	return codec.SliceToArray(codec.Address, addrs, ParamAllowedStateControllerAddresses)
}

func (e OutputAddressList) Decode(r dict.Dict) ([]*cryptolib.Address, error) {
	return codec.SliceFromArray(codec.Address, r, ParamAllowedStateControllerAddresses)
}

type OutputChainInfo struct{}

func (o OutputChainInfo) Encode(info *isc.ChainInfo) dict.Dict {
	ret := dict.Dict{
		ParamChainID:         codec.ChainID.Encode(info.ChainID),
		varChainOwnerID:      codec.AgentID.Encode(info.ChainOwnerID),
		varGasFeePolicyBytes: info.GasFeePolicy.Bytes(),
		varGasLimitsBytes:    info.GasLimits.Bytes(),
		varMetadata:          info.Metadata.Bytes(),
	}
	if len(info.PublicURL) > 0 {
		ret.Set(varPublicURL, codec.String.Encode(info.PublicURL))
	}
	return ret
}

func (o OutputChainInfo) Decode(r dict.Dict) (*isc.ChainInfo, error) {
	chainID, err := codec.ChainID.Decode(r[ParamChainID])
	if err != nil {
		return nil, err
	}
	return NewStateReader(r).GetChainInfo(chainID), nil
}

type InputAddCandidateNode struct{}

func (InputAddCandidateNode) Encode(a *AccessNodeInfo) dict.Dict {
	return dict.Dict{
		ParamAccessNodeInfoForCommittee: codec.Bool.Encode(a.ForCommittee),
		ParamAccessNodeInfoPubKey:       a.NodePubKey,
		ParamAccessNodeInfoCertificate:  a.Certificate,
		ParamAccessNodeInfoAccessAPI:    codec.String.Encode(a.AccessAPI),
	}
}

func (InputAddCandidateNode) Decode(d dict.Dict) (*AccessNodeInfo, error) {
	return &AccessNodeInfo{
		NodePubKey:   d[ParamAccessNodeInfoPubKey],
		Certificate:  d[ParamAccessNodeInfoCertificate],
		ForCommittee: lo.Must(codec.Bool.Decode(d[ParamAccessNodeInfoForCommittee], false)),
		AccessAPI:    lo.Must(codec.String.Decode(d[ParamAccessNodeInfoAccessAPI], "")),
	}, nil
}

type InputRevokeAccessNode struct{}

func (e InputRevokeAccessNode) Encode(a *AccessNodeInfo) dict.Dict {
	return dict.Dict{
		ParamAccessNodeInfoPubKey:      a.NodePubKey,
		ParamAccessNodeInfoCertificate: a.Certificate,
	}
}

func (InputRevokeAccessNode) Decode(d dict.Dict) (*AccessNodeInfo, error) {
	return &AccessNodeInfo{
		NodePubKey:  d[ParamAccessNodeInfoPubKey],
		Certificate: d[ParamAccessNodeInfoCertificate],
	}, nil
}

type InputChangeAccessNodes struct{}

func (InputChangeAccessNodes) Encode(r ChangeAccessNodesRequest) dict.Dict {
	d := dict.New()
	actionsMap := collections.NewMap(d, ParamChangeAccessNodesActions)
	for pubKey, action := range r {
		actionsMap.SetAt(pubKey[:], []byte{byte(action)})
	}
	return d
}

var errInvalidAction = coreerrors.Register("invalid action").Create()

func (InputChangeAccessNodes) Decode(d dict.Dict) (ChangeAccessNodesRequest, error) {
	actions := NewChangeAccessNodesRequest()
	m := collections.NewMapReadOnly(d, ParamChangeAccessNodesActions)
	var err error
	m.Iterate(func(pubKey, actionBin []byte) bool {
		var pk cryptolib.PublicKeyKey
		if len(pubKey) != len(pk) {
			err = errors.New("invalid public key")
			return false
		}
		copy(pk[:], pubKey)
		var action byte
		action, err = codec.Uint8.Decode(actionBin)
		if err != nil || action >= byte(ChangeAccessNodeActionLast) {
			err = errInvalidAction
			return false
		}
		actions[pk] = ChangeAccessNodeAction(action)
		return true
	})
	return actions, err
}

type OutputChainNodes struct{}

func (OutputChainNodes) Encode(r *GetChainNodesResponse) dict.Dict {
	res := dict.New()
	candidates := collections.NewMap(res, ParamGetChainNodesAccessNodeCandidates)
	for pk, ani := range r.AccessNodeCandidates {
		candidates.SetAt(pk[:], ani.Bytes())
	}
	nodes := collections.NewMap(res, ParamGetChainNodesAccessNodes)
	for pk := range r.AccessNodes {
		nodes.SetAt(pk[:], []byte{0x01})
	}
	return res
}

func (OutputChainNodes) Decode(d dict.Dict) (*GetChainNodesResponse, error) {
	res := &GetChainNodesResponse{
		AccessNodeCandidates: make(map[cryptolib.PublicKeyKey]*AccessNodeInfo),
		AccessNodes:          make(map[cryptolib.PublicKeyKey]struct{}),
	}

	var err error
	ac := collections.NewMapReadOnly(d, ParamGetChainNodesAccessNodeCandidates)
	ac.Iterate(func(pubKey, value []byte) bool {
		var ani *AccessNodeInfo
		ani, err = AccessNodeInfoFromBytes(pubKey, value)
		if err != nil {
			return false
		}
		var pk *cryptolib.PublicKey
		pk, err = cryptolib.PublicKeyFromBytes(pubKey)
		if err != nil {
			return false
		}
		res.AccessNodeCandidates[pk.AsKey()] = ani
		return true
	})
	if err != nil {
		return nil, err
	}

	an := collections.NewMapReadOnly(d, ParamGetChainNodesAccessNodes)
	an.Iterate(func(pubKeyBin, value []byte) bool {
		var pk *cryptolib.PublicKey
		pk, err = cryptolib.PublicKeyFromBytes(pubKeyBin)
		if err != nil {
			return false
		}
		res.AccessNodes[pk.AsKey()] = struct{}{}
		return true
	})
	return res, err
}
