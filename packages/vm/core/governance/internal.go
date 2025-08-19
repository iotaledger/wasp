// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package governance

import (
	"math/big"

	"github.com/samber/lo"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/packages/coin"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/kv/codec"
	"github.com/iotaledger/wasp/v2/packages/kv/collections"
	"github.com/iotaledger/wasp/v2/packages/parameters"
	"github.com/iotaledger/wasp/v2/packages/vm/gas"
)

func (s *StateWriter) SetInitialState(chainAdmin isc.AgentID, blockKeepAmount int32) {
	s.SetChainAdmin(chainAdmin)
	s.SetGasFeePolicy(gas.DefaultFeePolicy())
	s.SetGasLimits(gas.LimitsDefault)
	s.SetMaintenanceStatus(false)
	s.SetBlockKeepAmount(blockKeepAmount)
	s.SetGasCoinTargetValue(isc.GasCoinTargetValue)
	s.SetPayoutAgentID(chainAdmin)
}

// GetChainInfo returns global variables of the chain
func (s *StateReader) GetChainInfo(chainID isc.ChainID) *isc.ChainInfo {
	ret := &isc.ChainInfo{
		Metadata: &isc.PublicChainMetadata{},
		ChainID:  chainID,
	}
	ret.ChainAdmin = s.GetChainAdmin()
	ret.GasFeePolicy = s.GetGasFeePolicy()
	ret.GasLimits = s.GetGasLimits()
	ret.BlockKeepAmount = s.GetBlockKeepAmount()
	ret.PublicURL = s.GetPublicURL()
	ret.Metadata = s.GetMetadata()
	return ret
}

func (s *StateReader) GetGasCoinTargetValue() coin.Value {
	return lo.Must(codec.Decode[coin.Value](s.state.Get(varGasCoinTargetValue)))
}

func (s *StateWriter) SetGasCoinTargetValue(m coin.Value) {
	s.state.Set(varGasCoinTargetValue, codec.Encode(m))
}

func (s *StateReader) GetChainAdmin() isc.AgentID {
	return lo.Must(codec.Decode[isc.AgentID](s.state.Get(varChainAdmin)))
}

func (s *StateWriter) SetChainAdmin(a isc.AgentID) {
	s.state.Set(varChainAdmin, codec.Encode(a))
	if s.GetChainAdminDelegated() != nil {
		s.state.Del(varChainAdminDelegated)
	}
}

func (s *StateReader) GetChainAdminDelegated() isc.AgentID {
	return lo.Must(codec.Decode[isc.AgentID](s.state.Get(varChainAdminDelegated), nil))
}

func (s *StateWriter) SetChainAdminDelegated(a isc.AgentID) {
	s.state.Set(varChainAdminDelegated, codec.Encode(a))
}

func (s *StateReader) GetPayoutAgentID() isc.AgentID {
	return lo.Must(codec.Decode[isc.AgentID](s.state.Get(varPayoutAgentID)))
}

func (s *StateWriter) SetPayoutAgentID(a isc.AgentID) {
	s.state.Set(varPayoutAgentID, codec.Encode(a))
}

func (s *StateReader) GetGasFeePolicy() *gas.FeePolicy {
	return lo.Must(gas.FeePolicyFromBytes(s.state.Get(varGasFeePolicyBytes)))
}

func (s *StateWriter) SetGasFeePolicy(fp *gas.FeePolicy) {
	s.state.Set(varGasFeePolicyBytes, fp.Bytes())
}

func (s *StateReader) GetDefaultGasPrice() *big.Int {
	return s.GetGasFeePolicy().DefaultGasPriceFullDecimals(parameters.BaseTokenDecimals)
}

func (s *StateReader) GetGasLimits() *gas.Limits {
	data := s.state.Get(varGasLimitsBytes)
	if data == nil {
		return gas.LimitsDefault
	}
	return lo.Must(gas.LimitsFromBytes(data))
}

func (s *StateWriter) SetGasLimits(gl *gas.Limits) {
	s.state.Set(varGasLimitsBytes, gl.Bytes())
}

func (s *StateReader) GetBlockKeepAmount() int32 {
	return lo.Must(codec.Decode[int32](s.state.Get(varBlockKeepAmount), DefaultBlockKeepAmount))
}

func (s *StateWriter) SetBlockKeepAmount(n int32) {
	s.state.Set(varBlockKeepAmount, codec.Encode(n))
}

func (s *StateWriter) SetPublicURL(url string) {
	s.state.Set(varPublicURL, codec.Encode(url))
}

func (s *StateReader) GetPublicURL() string {
	return codec.MustDecode(s.state.Get(varPublicURL), "")
}

func (s *StateWriter) SetMetadata(metadata *isc.PublicChainMetadata) {
	s.state.Set(varMetadata, metadata.Bytes())
}

func (s *StateReader) GetMetadata() *isc.PublicChainMetadata {
	metadataBytes := s.state.Get(varMetadata)
	if metadataBytes == nil {
		return &isc.PublicChainMetadata{}
	}
	return lo.Must(isc.PublicChainMetadataFromBytes(metadataBytes))
}

func (s *StateWriter) AccessNodesMap() *collections.Map {
	return collections.NewMap(s.state, varAccessNodes)
}

func (s *StateReader) AccessNodesMap() *collections.ImmutableMap {
	return collections.NewMapReadOnly(s.state, varAccessNodes)
}

func (s *StateWriter) AccessNodeCandidatesMap() *collections.Map {
	return collections.NewMap(s.state, varAccessNodeCandidates)
}

func (s *StateReader) AccessNodeCandidatesMap() *collections.ImmutableMap {
	return collections.NewMapReadOnly(s.state, varAccessNodeCandidates)
}

func (s *StateReader) GetMaintenanceStatus() bool {
	r := s.state.Get(varMaintenanceStatus)
	if r == nil {
		return false // chain is being initialized, governance has not been initialized yet
	}
	return lo.Must(codec.Decode[bool](r))
}

func (s *StateWriter) SetMaintenanceStatus(status bool) {
	s.state.Set(varMaintenanceStatus, codec.Encode(status))
}

func (s *StateReader) AccessNodes() []*cryptolib.PublicKey {
	accessNodes := []*cryptolib.PublicKey{}
	s.AccessNodesMap().IterateKeys(func(pubKeyBytes []byte) bool {
		pubKey := lo.Must(cryptolib.PublicKeyFromBytes(pubKeyBytes))
		accessNodes = append(accessNodes, pubKey)
		return true
	})
	return accessNodes
}

func (s *StateReader) CandidateNodes() []*AccessNodeInfo {
	candidateNodes := []*AccessNodeInfo{}
	s.AccessNodeCandidatesMap().Iterate(func(pubKeyBytes, accessNodeDataBytes []byte) bool {
		pubKey := lo.Must(cryptolib.PublicKeyFromBytes(pubKeyBytes))
		and := bcs.MustUnmarshal[AccessNodeData](accessNodeDataBytes)
		candidateNodes = append(candidateNodes, &AccessNodeInfo{
			NodePubKey:     pubKey,
			AccessNodeData: and,
		})
		return true
	})
	return candidateNodes
}
