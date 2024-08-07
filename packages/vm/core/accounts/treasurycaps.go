package accounts

import (
	"github.com/samber/lo"

	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
)

func treasuryCapsMapKey(agentID isc.AgentID) string {
	return prefixTreasuryCaps + string(agentID.Bytes())
}

func (s *StateWriter) accountTreasuryCapsMap(agentID isc.AgentID) *collections.Map {
	return collections.NewMap(s.state, treasuryCapsMapKey(agentID))
}

func (s *StateReader) accountTreasuryCapsMapR(agentID isc.AgentID) *collections.ImmutableMap {
	return collections.NewMapReadOnly(s.state, treasuryCapsMapKey(agentID))
}

func (s *StateWriter) allTreasuryCapsMap() *collections.Map {
	return collections.NewMap(s.state, keyTreasuryCapRecords)
}

func (s *StateReader) allTreasuryCapsMapR() *collections.ImmutableMap {
	return collections.NewMapReadOnly(s.state, keyTreasuryCapRecords)
}

func (s *StateWriter) SaveTreasuryCap(rec *TreasuryCapRecord) {
	m := s.allTreasuryCapsMap()
	key := rec.CoinType.Bytes()
	if m.HasAt(key) {
		panic(ErrDuplicateTreasuryCap)
	}
	m.SetAt(key, rec.Bytes())
}

func (s *StateWriter) DeleteTreasuryCap(coinType coin.Type) {
	s.allTreasuryCapsMap().DelAt(coinType.Bytes())
}

func (s *StateReader) GetTreasuryCap(coinType coin.Type, chainID isc.ChainID) *TreasuryCapRecord {
	data := s.allTreasuryCapsMapR().GetAt(coinType.Bytes())
	if data == nil {
		return nil
	}
	return lo.Must(TreasuryCapRecordFromBytes(data, coinType))
}

func (s *StateReader) hasTreasuryCap(agentID isc.AgentID, coinType coin.Type) bool {
	return s.accountTreasuryCapsMapR(agentID).HasAt(coinType.Bytes())
}

func (s *StateWriter) addTreasuryCapToAccount(agentID isc.AgentID, coinType coin.Type) {
	key := codec.CoinType.Encode(coinType)
	treasuryCaps := s.accountTreasuryCapsMap(agentID)
	if treasuryCaps.HasAt(key) {
		panic(ErrDuplicateTreasuryCap)
	}
	treasuryCaps.SetAt(key, codec.Bool.Encode(true))
}

func (s *StateWriter) deleteTreasuryCapFromAccount(agentID isc.AgentID, coinType coin.Type) {
	key := codec.CoinType.Encode(coinType)
	treasuryCaps := s.accountTreasuryCapsMap(agentID)
	if !treasuryCaps.HasAt(key) {
		panic(ErrTreasuryCapNotFound)
	}
	treasuryCaps.DelAt(key)
}

func (s *StateWriter) MoveTreasuryCapBetweenAccounts(agentIDFrom, agentIDTo isc.AgentID, coinType coin.Type) {
	s.deleteTreasuryCapFromAccount(agentIDFrom, coinType)
	s.addTreasuryCapToAccount(agentIDTo, coinType)
}
