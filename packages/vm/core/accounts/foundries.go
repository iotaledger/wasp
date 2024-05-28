package accounts

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
)

func (s *StateWriter) newFoundriesArray() *collections.Array {
	return collections.NewArray(s.state, keyNewFoundries)
}

func (s *StateWriter) accountFoundriesMap(agentID isc.AgentID) *collections.Map {
	return collections.NewMap(s.state, foundriesMapKey(agentID))
}

func (s *StateReader) accountFoundriesMapR(agentID isc.AgentID) *collections.ImmutableMap {
	return collections.NewMapReadOnly(s.state, foundriesMapKey(agentID))
}

func (s *StateWriter) allFoundriesMap() *collections.Map {
	return collections.NewMap(s.state, keyFoundryOutputRecords)
}

func (s *StateReader) allFoundriesMapR() *collections.ImmutableMap {
	return collections.NewMapReadOnly(s.state, keyFoundryOutputRecords)
}

// SaveFoundryOutput stores foundry output into the map of all foundry outputs (compressed form)
func (s *StateWriter) SaveFoundryOutput(f *iotago.FoundryOutput, outputIndex uint16) {
	foundryRec := foundryOutputRec{
		// TransactionID is unknown yet, will be filled next block
		OutputID:    iotago.OutputIDFromTransactionIDAndIndex(iotago.TransactionID{}, outputIndex),
		Amount:      f.Amount,
		TokenScheme: f.TokenScheme,
		Metadata:    []byte{},
	}

	if f.FeatureSet().MetadataFeature() != nil {
		foundryRec.Metadata = f.FeatureSet().MetadataFeature().Data
	}

	s.allFoundriesMap().SetAt(codec.Uint32.Encode(f.SerialNumber), foundryRec.Bytes())
	s.newFoundriesArray().Push(codec.Uint32.Encode(f.SerialNumber))
}

func (s *StateWriter) updateFoundryOutputIDs(anchorTxID iotago.TransactionID) {
	newFoundries := s.newFoundriesArray()
	allFoundries := s.allFoundriesMap()
	n := newFoundries.Len()
	for i := uint32(0); i < n; i++ {
		k := newFoundries.GetAt(i)
		rec := mustFoundryOutputRecFromBytes(allFoundries.GetAt(k))
		rec.OutputID = iotago.OutputIDFromTransactionIDAndIndex(anchorTxID, rec.OutputID.Index())
		allFoundries.SetAt(k, rec.Bytes())
	}
	newFoundries.Erase()
}

// DeleteFoundryOutput deletes foundry output from the map of all foundries
func (s *StateWriter) DeleteFoundryOutput(sn uint32) {
	s.allFoundriesMap().DelAt(codec.Uint32.Encode(sn))
}

// GetFoundryOutput returns foundry output, its block number and output index
func (s *StateReader) GetFoundryOutput(sn uint32, chainID isc.ChainID) (*iotago.FoundryOutput, iotago.OutputID) {
	data := s.allFoundriesMapR().GetAt(codec.Uint32.Encode(sn))
	if data == nil {
		return nil, iotago.OutputID{}
	}
	rec := mustFoundryOutputRecFromBytes(data)

	ret := &iotago.FoundryOutput{
		Amount:       rec.Amount,
		NativeTokens: nil,
		SerialNumber: sn,
		TokenScheme:  rec.TokenScheme,
		Conditions: iotago.UnlockConditions{
			&iotago.ImmutableAliasUnlockCondition{Address: chainID.AsAddress().AsIotagoAddress().(*iotago.AliasAddress)},
		},
		Features: nil,
	}

	if len(rec.Metadata) > 0 {
		ret.Features = []iotago.Feature{
			&iotago.MetadataFeature{Data: rec.Metadata},
		}
	}

	return ret, rec.OutputID
}

// hasFoundry checks if specific account owns the foundry
func (s *StateReader) hasFoundry(agentID isc.AgentID, sn uint32) bool {
	return s.accountFoundriesMapR(agentID).HasAt(codec.Uint32.Encode(sn))
}

// addFoundryToAccount adds new foundry to the foundries controlled by the account
func (s *StateWriter) addFoundryToAccount(agentID isc.AgentID, sn uint32) {
	key := codec.Uint32.Encode(sn)
	foundries := s.accountFoundriesMap(agentID)
	if foundries.HasAt(key) {
		panic(ErrRepeatingFoundrySerialNumber)
	}
	foundries.SetAt(key, codec.Bool.Encode(true))
}

func (s *StateWriter) deleteFoundryFromAccount(agentID isc.AgentID, sn uint32) {
	key := codec.Uint32.Encode(sn)
	foundries := s.accountFoundriesMap(agentID)
	if !foundries.HasAt(key) {
		panic(ErrFoundryNotFound)
	}
	foundries.DelAt(key)
}

// MoveFoundryBetweenAccounts changes ownership of the foundry
func (s *StateWriter) MoveFoundryBetweenAccounts(agentIDFrom, agentIDTo isc.AgentID, sn uint32) {
	s.deleteFoundryFromAccount(agentIDFrom, sn)
	s.addFoundryToAccount(agentIDTo, sn)
}
