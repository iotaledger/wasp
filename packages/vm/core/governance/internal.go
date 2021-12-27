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

// MustGetChainInfo return global variables of the chain
func MustGetChainInfo(state kv.KVStoreReader) *ChainInfo {
	d := kvdecoder.New(state)
	ret := &ChainInfo{
		ChainID:         d.MustGetChainID(VarChainID),
		ChainOwnerID:    d.MustGetAgentID(VarChainOwnerID),
		Description:     d.MustGetString(VarDescription, ""),
		GasFeePolicy:    GetGasFeePolicy(state),
		MaxBlobSize:     d.MustGetUint32(VarMaxBlobSize, 0),
		MaxEventSize:    d.MustGetUint16(VarMaxEventSize, 0),
		MaxEventsPerReq: d.MustGetUint16(VarMaxEventsPerReq, 0),
	}
	return ret
}

func MustGetChainOwnerID(state kv.KVStoreReader) *iscp.AgentID {
	d := kvdecoder.New(state)
	return d.MustGetAgentID(VarChainOwnerID)
}

// GetGasFeePolicy returns gas policy from the state
func GetGasFeePolicy(state kv.KVStoreReader) *gas.GasFeePolicy {
	return gas.MustGasFeePolicyFromBytes(state.MustGet(VarGasFeePolicyBytes))
}
