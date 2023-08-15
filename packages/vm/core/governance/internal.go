// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package governance

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

// GetRotationAddress tries to read the state of 'governance' and extract rotation address
// If succeeds, it means this block is fake.
// If fails, return nil
func GetRotationAddress(state kv.KVStoreReader) iotago.Address {
	ret, err := codec.DecodeAddress(state.Get(VarRotateToAddress), nil)
	if err != nil {
		return nil
	}
	return ret
}

// GetChainInfo returns global variables of the chain
func GetChainInfo(state kv.KVStoreReader, chainID isc.ChainID) (*isc.ChainInfo, error) {
	d := kvdecoder.New(state)
	ret := &isc.ChainInfo{
		ChainID:  chainID,
		Metadata: &isc.PublicChainMetadata{},
	}
	var err error
	if ret.ChainOwnerID, err = d.GetAgentID(VarChainOwnerID); err != nil {
		return nil, err
	}

	if ret.GasFeePolicy, err = GetGasFeePolicy(state); err != nil {
		return nil, err
	}

	if ret.GasLimits, err = GetGasLimits(state); err != nil {
		return nil, err
	}

	ret.BlockKeepAmount = GetBlockKeepAmount(state)
	if ret.PublicURL, err = GetPublicURL(state); err != nil {
		return nil, err
	}

	if ret.Metadata, err = GetMetadata(state); err != nil {
		return nil, err
	}

	return ret, nil
}

// MustGetChainInfo return global variables of the chain
func MustGetChainInfo(state kv.KVStoreReader, chainID isc.ChainID) *isc.ChainInfo {
	info, err := GetChainInfo(state, chainID)
	if err != nil {
		panic(err)
	}
	return info
}

func MustGetMinCommonAccountBalance(state kv.KVStoreReader) uint64 {
	return kvdecoder.New(state).MustGetUint64(VarMinBaseTokensOnCommonAccount)
}

func MustGetPayoutAgentID(state kv.KVStoreReader) isc.AgentID {
	return kvdecoder.New(state).MustGetAgentID(VarPayoutAgentID)
}

func mustGetChainOwnerID(state kv.KVStoreReader) isc.AgentID {
	d := kvdecoder.New(state)
	return d.MustGetAgentID(VarChainOwnerID)
}

// GetGasFeePolicy returns gas policy from the state
func GetGasFeePolicy(state kv.KVStoreReader) (*gas.FeePolicy, error) {
	return gas.FeePolicyFromBytes(state.Get(VarGasFeePolicyBytes))
}

func MustGetGasFeePolicy(state kv.KVStoreReader) *gas.FeePolicy {
	return gas.MustFeePolicyFromBytes(state.Get(VarGasFeePolicyBytes))
}

func MustGetGasLimits(state kv.KVStoreReader) *gas.Limits {
	gl, err := GetGasLimits(state)
	if err != nil {
		panic(err)
	}
	return gl
}

func GetGasLimits(state kv.KVStoreReader) (*gas.Limits, error) {
	data := state.Get(VarGasLimitsBytes)
	if data == nil {
		return gas.LimitsDefault, nil
	}
	return gas.LimitsFromBytes(data)
}

func GetBlockKeepAmount(state kv.KVStoreReader) int32 {
	return codec.MustDecodeInt32(state.Get(VarBlockKeepAmount), DefaultBlockKeepAmount)
}

func SetPublicURL(state kv.KVStore, url string) {
	state.Set(VarPublicURL, codec.EncodeString(url))
}

func GetPublicURL(state kv.KVStoreReader) (string, error) {
	return codec.DecodeString(state.Get(VarPublicURL), "")
}

func SetMetadata(state kv.KVStore, metadata *isc.PublicChainMetadata) {
	state.Set(VarMetadata, metadata.Bytes())
}

func GetMetadata(state kv.KVStoreReader) (*isc.PublicChainMetadata, error) {
	metadataBytes := state.Get(VarMetadata)
	if metadataBytes == nil {
		return &isc.PublicChainMetadata{}, nil
	}
	return isc.PublicChainMetadataFromBytes(metadataBytes)
}

func MustGetMetadata(state kv.KVStoreReader) *isc.PublicChainMetadata {
	metadata, err := GetMetadata(state)
	if err != nil {
		panic(err)
	}
	return metadata
}

func AccessNodesMap(state kv.KVStore) *collections.Map {
	return collections.NewMap(state, VarAccessNodes)
}

func AccessNodesMapR(state kv.KVStoreReader) *collections.ImmutableMap {
	return collections.NewMapReadOnly(state, VarAccessNodes)
}

func AccessNodeCandidatesMap(state kv.KVStore) *collections.Map {
	return collections.NewMap(state, VarAccessNodeCandidates)
}

func AccessNodeCandidatesMapR(state kv.KVStoreReader) *collections.ImmutableMap {
	return collections.NewMapReadOnly(state, VarAccessNodeCandidates)
}
