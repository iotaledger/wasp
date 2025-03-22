package accounts

import (
	"github.com/samber/lo"

	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/parameters"
)

func (s *StateWriter) coinInfosMap() *collections.Map {
	return collections.NewMap(s.state, KeyCoinInfo)
}

func (s *StateReader) coinInfosMapR() *collections.ImmutableMap {
	return collections.NewMapReadOnly(s.state, KeyCoinInfo)
}

func (s *StateWriter) SaveCoinInfo(rec *parameters.IotaCoinInfo) {
	s.coinInfosMap().SetAt(rec.CoinType.Bytes(), rec.Bytes())
}

func (s *StateWriter) DeleteCoinInfo(coinType coin.Type) {
	s.coinInfosMap().DelAt(coinType.Bytes())
}

func (s *StateReader) GetCoinInfo(coinType coin.Type) (*parameters.IotaCoinInfo, bool) {
	data := s.coinInfosMapR().GetAt(coinType.Bytes())
	if data == nil {
		return nil, false
	}
	return lo.Must(parameters.IotaCoinInfoFromBytes(data)), true
}
