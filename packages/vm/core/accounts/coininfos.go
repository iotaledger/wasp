package accounts

import (
	"github.com/samber/lo"

	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/collections"
)

func (s *StateWriter) coinInfosMap() *collections.Map {
	return collections.NewMap(s.state, keyCoinInfo)
}

func (s *StateReader) coinInfosMapR() *collections.ImmutableMap {
	return collections.NewMapReadOnly(s.state, keyCoinInfo)
}

func (s *StateWriter) SaveCoinInfo(rec *isc.SuiCoinInfo) {
	s.coinInfosMap().SetAt(rec.CoinType.Bytes(), rec.Bytes())
}

func (s *StateWriter) DeleteCoinInfo(coinType coin.Type) {
	s.coinInfosMap().DelAt(coinType.Bytes())
}

func (s *StateReader) GetCoinInfo(coinType coin.Type) (*isc.SuiCoinInfo, bool) {
	data := s.coinInfosMapR().GetAt(coinType.Bytes())
	if data == nil {
		return nil, false
	}
	return lo.Must(isc.SuiCoinInfoFromBytes(data)), true
}
