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
	ret, err := codec.DecodeAddress(state.Get(StateVarRotateToAddress), nil)
	if err != nil {
		return nil
	}
	return ret
}

// GetChainInfo returns global variables of the chain
func GetChainInfo(state kv.KVStoreReader, chainID isc.ChainID) (*isc.ChainInfo, error) {
	d := kvdecoder.New(state)
	ret := &isc.ChainInfo{
		ChainID: chainID,
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
	if ret.PublicURL, err = GetPublicURL(state); err != nil {
		return nil, err
	}
	if ret.MetadataEVMJsonRPCURL, err = GetEVMJsonRPCURL(state); err != nil {
		return nil, err
	}
	if ret.MetadataEVMWebSocketURL, err = GetEVMWebSocketURL(state); err != nil {
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

func MustGetChainOwnerID(state kv.KVStoreReader) isc.AgentID {
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

func SetPublicURL(state kv.KVStore, url string) {
	state.Set(VarPublicURL, codec.EncodeString(url))
}

func GetPublicURL(state kv.KVStoreReader) (string, error) {
	return codec.DecodeString(state.Get(VarPublicURL), "")
}

func SetEVMJsonRPCURL(state kv.KVStore, url string) {
	state.Set(VarMetadataEVMJsonRPCURL, codec.EncodeString(url))
}

func GetEVMJsonRPCURL(state kv.KVStoreReader) (string, error) {
	return codec.DecodeString(state.Get(VarMetadataEVMJsonRPCURL), "")
}

func SetEVMWebSocketURL(state kv.KVStore, url string) {
	state.Set(VarMetadataEVMWebSocketURL, codec.EncodeString(url))
}

func GetEVMWebSocketURL(state kv.KVStoreReader) (string, error) {
	return codec.DecodeString(state.Get(VarMetadataEVMWebSocketURL), "")
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
