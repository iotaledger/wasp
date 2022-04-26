// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package governance

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

// GetRotationAddress tries to read the state of 'governance' and extract rotation address
// If succeeds, it means this block is fake.
// If fails, return nil
func GetRotationAddress(state kv.KVStoreReader) iotago.Address {
	ret, err := codec.DecodeAddress(state.MustGet(StateVarRotateToAddress), nil)
	if err != nil {
		return nil
	}
	return ret
}

// GetChainInfo returns global variables of the chain
func GetChainInfo(state kv.KVStoreReader) (*ChainInfo, error) {
	d := kvdecoder.New(state)
	ret := &ChainInfo{}
	var err error
	if ret.ChainID, err = d.GetChainID(VarChainID); err != nil {
		return nil, err
	}
	if ret.ChainOwnerID, err = d.GetAgentID(VarChainOwnerID); err != nil {
		return nil, err
	}
	if ret.Description, err = d.GetString(VarDescription, ""); err != nil {
		return nil, err
	}
	if ret.GasFeePolicy, err = GetGasFeePolicy(state); err != nil {
		return nil, err
	}
	if ret.MaxBlobSize, err = d.GetUint32(VarMaxBlobSize, 0); err != nil {
		return nil, err
	}
	if ret.MaxEventSize, err = d.GetUint16(VarMaxEventSize, 0); err != nil {
		return nil, err
	}
	if ret.MaxEventsPerReq, err = d.GetUint16(VarMaxEventsPerReq, 0); err != nil {
		return nil, err
	}
	return ret, nil
}

// MustGetChainInfo return global variables of the chain
func MustGetChainInfo(state kv.KVStoreReader) *ChainInfo {
	info, err := GetChainInfo(state)
	if err != nil {
		panic(err)
	}
	return info
}

func MustGetChainOwnerID(state kv.KVStoreReader) *iscp.AgentID {
	d := kvdecoder.New(state)
	return d.MustGetAgentID(VarChainOwnerID)
}

// GetGasFeePolicy returns gas policy from the state
func GetGasFeePolicy(state kv.KVStoreReader) (*gas.GasFeePolicy, error) {
	return gas.FeePolicyFromBytes(state.MustGet(VarGasFeePolicyBytes))
}

func MustGetGasFeePolicy(state kv.KVStoreReader) *gas.GasFeePolicy {
	return gas.MustGasFeePolicyFromBytes(state.MustGet(VarGasFeePolicyBytes))
}
