package isc

import (
	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/packages/coin"
)

type IotaCoinInfo struct {
	CoinType    coin.Type
	Decimals    uint8
	Name        string
	Symbol      string
	Description string
	IconURL     string
	// TotalSupply is a dynamic value, it increases every epoch
	TotalSupply coin.Value
}

type IotaCoinInfos = map[coin.Type]*IotaCoinInfo

func (s *IotaCoinInfo) Bytes() []byte {
	return bcs.MustMarshal(s)
}

func (s *IotaCoinInfo) Equals(other *IotaCoinInfo) bool {
	return s.CoinType == other.CoinType &&
		s.Decimals == other.Decimals &&
		s.Name == other.Name &&
		s.Symbol == other.Symbol &&
		s.Description == other.Description &&
		s.IconURL == other.IconURL &&
		s.TotalSupply == other.TotalSupply
}

func IotaCoinInfoFromBytes(b []byte) (*IotaCoinInfo, error) {
	ret, err := bcs.Unmarshal[IotaCoinInfo](b)
	return &ret, err
}

func IotaCoinInfoFromL1Metadata(
	coinType coin.Type,
	metadata *iotajsonrpc.IotaCoinMetadata,
	totalSupply coin.Value,
) *IotaCoinInfo {
	return &IotaCoinInfo{
		CoinType:    coinType,
		Decimals:    metadata.Decimals,
		Name:        metadata.Name,
		Symbol:      metadata.Symbol,
		Description: metadata.Description,
		IconURL:     metadata.IconUrl,
		TotalSupply: totalSupply,
	}
}

var BaseTokenCoinInfo = &IotaCoinInfo{CoinType: coin.BaseTokenType}
