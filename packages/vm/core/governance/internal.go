// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package governance

import (
	"math/big"

	"github.com/samber/lo"

	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/util/bcs"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

func (s *StateWriter) SetInitialState(chainOwner isc.AgentID, blockKeepAmount int32) {
	s.SetChainOwnerID(chainOwner)
	s.SetGasFeePolicy(gas.DefaultFeePolicy())
	s.SetGasLimits(gas.LimitsDefault)
	s.SetMaintenanceStatus(false)
	s.SetBlockKeepAmount(blockKeepAmount)
	s.SetGasCoinTargetValue(isc.GasCoinTargetValue)
	s.SetPayoutAgentID(chainOwner)
}

// GetRotationAddress tries to read the state of 'governance' and extract rotation address
// If succeeds, it means this block is fake.
// If fails, return nil
func (s *StateReader) GetRotationAddress() *cryptolib.Address {
	ret, err := codec.Decode[*cryptolib.Address](s.state.Get(varRotateToAddress), nil)
	if err != nil {
		return nil
	}
	return ret
}

func (s *StateWriter) SetRotationAddress(a *cryptolib.Address) {
	s.state.Set(varRotateToAddress, codec.Encode(a))
}

// GetChainInfo returns global variables of the chain
func (s *StateReader) GetChainInfo(chainID isc.ChainID) *isc.ChainInfo {
	ret := &isc.ChainInfo{
		ChainID:  chainID,
		Metadata: &isc.PublicChainMetadata{},
	}
	ret.ChainOwnerID = s.GetChainOwnerID()
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

func (s *StateReader) GetChainOwnerID() isc.AgentID {
	return lo.Must(codec.Decode[isc.AgentID](s.state.Get(VarChainOwnerID)))
}

func (s *StateWriter) SetChainOwnerID(a isc.AgentID) {
	s.state.Set(VarChainOwnerID, codec.Encode(a))
	if s.GetChainOwnerIDDelegated() != nil {
		s.state.Del(VarChainOwnerIDDelegated)
	}
}

func (s *StateReader) GetChainOwnerIDDelegated() isc.AgentID {
	return lo.Must(codec.Decode[isc.AgentID](s.state.Get(VarChainOwnerIDDelegated), nil))
}

func (s *StateWriter) SetChainOwnerIDDelegated(a isc.AgentID) {
	s.state.Set(VarChainOwnerIDDelegated, codec.Encode(a))
}

func (s *StateReader) GetPayoutAgentID() isc.AgentID {
	return lo.Must(codec.Decode[isc.AgentID](s.state.Get(VarPayoutAgentID)))
}

func (s *StateWriter) SetPayoutAgentID(a isc.AgentID) {
	s.state.Set(VarPayoutAgentID, codec.Encode(a))
}

func (s *StateReader) GetGasFeePolicy() *gas.FeePolicy {
	return lo.Must(gas.FeePolicyFromBytes(s.state.Get(varGasFeePolicyBytes)))
}

func (s *StateWriter) SetGasFeePolicy(fp *gas.FeePolicy) {
	s.state.Set(varGasFeePolicyBytes, fp.Bytes())
}

func (s *StateReader) GetDefaultGasPrice() *big.Int {
	return s.GetGasFeePolicy().DefaultGasPriceFullDecimals(parameters.Decimals)
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

func (s *StateWriter) AllowedStateControllerAddressesMap() *collections.Map {
	return collections.NewMap(s.state, varAllowedStateControllerAddresses)
}

func (s *StateReader) AllowedStateControllerAddressesMap() *collections.ImmutableMap {
	return collections.NewMapReadOnly(s.state, varAllowedStateControllerAddresses)
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
