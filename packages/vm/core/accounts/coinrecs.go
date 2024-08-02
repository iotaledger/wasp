package accounts

import (
	"github.com/samber/lo"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/collections"
)

func (s *StateWriter) coinRecordsMap() *collections.Map {
	return collections.NewMap(s.state, keyCoinRecords)
}

func (s *StateReader) coinRecordsMapR() *collections.ImmutableMap {
	return collections.NewMapReadOnly(s.state, keyCoinRecords)
}

func (s *StateWriter) SaveCoin(rec CoinRecord) {
	s.coinRecordsMap().SetAt(rec.CoinType.Bytes(), rec.Bytes())
}

func (s *StateWriter) DeleteCoin(coinType isc.CoinType) {
	s.coinRecordsMap().DelAt(coinType.Bytes())
}

func (s *StateReader) GetCoin(coinType isc.CoinType, chainID isc.ChainID) *CoinRecord {
	data := s.coinRecordsMapR().GetAt(coinType.Bytes())
	if data == nil {
		return nil
	}
	return lo.Must(CoinRecordFromBytes(data, coinType))
}
