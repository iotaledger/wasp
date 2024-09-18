package isc

import (
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/util/bcs"
	"github.com/iotaledger/wasp/sui-go/suijsonrpc"
)

type SuiCoinInfo struct {
	CoinType    coin.Type
	Decimals    uint8
	Name        string
	Symbol      string
	Description string
	IconURL     string
	TotalSupply coin.Value
}

type SuiCoinInfos = map[coin.Type]*SuiCoinInfo

func (s *SuiCoinInfo) Bytes() []byte {
	return bcs.MustMarshal(s)
}

func (s *SuiCoinInfo) Equals(other *SuiCoinInfo) bool {
	return s.CoinType == other.CoinType &&
		s.Decimals == other.Decimals &&
		s.Name == other.Name &&
		s.Symbol == other.Symbol &&
		s.Description == other.Description &&
		s.IconURL == other.IconURL &&
		s.TotalSupply == other.TotalSupply
}

func SuiCoinInfoFromBytes(b []byte) (*SuiCoinInfo, error) {
	ret, err := bcs.Unmarshal[SuiCoinInfo](b)
	return &ret, err
}

func SuiCoinInfoFromL1Metadata(
	coinType coin.Type,
	metadata *suijsonrpc.SuiCoinMetadata,
	totalSupply coin.Value,
) *SuiCoinInfo {
	return &SuiCoinInfo{
		CoinType:    coinType,
		Decimals:    metadata.Decimals,
		Name:        metadata.Name,
		Symbol:      metadata.Symbol,
		Description: metadata.Description,
		IconURL:     metadata.IconUrl,
		TotalSupply: totalSupply,
	}
}
