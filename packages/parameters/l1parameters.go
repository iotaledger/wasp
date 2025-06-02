// Package parameters provides configuration parameters for L1 and other components of the system.
package parameters

import (
	"encoding/json"

	"github.com/samber/lo"

	bcs "github.com/iotaledger/bcs-go"

	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/hashing"
)

const BaseTokenDecimals = 9

// L1Params describes parameters coming from the L1Params node
type L1Params struct {
	Protocol  *Protocol     `json:"protocol" swagger:"required"`
	BaseToken *IotaCoinInfo `json:"baseToken" swagger:"required"`
}

func (l *L1Params) String() string {
	return string(lo.Must(json.MarshalIndent(l, "", "  ")))
}

func (l *L1Params) Bytes() []byte {
	b, err := bcs.Marshal(l)
	if err != nil {
		panic(err)
	}
	return b
}

func (l *L1Params) Hash() hashing.HashValue {
	return hashing.HashData(l.Bytes())
}

type Protocol struct {
	Epoch                 *iotajsonrpc.BigInt `json:"epoch" swagger:"required"`
	ProtocolVersion       *iotajsonrpc.BigInt `json:"protocol_version" swagger:"required"`
	SystemStateVersion    *iotajsonrpc.BigInt `json:"system_state_version" swagger:"required"`
	ReferenceGasPrice     *iotajsonrpc.BigInt `json:"reference_gas_price" swagger:"required"`
	EpochStartTimestampMs *iotajsonrpc.BigInt `json:"epoch_start_timestamp_ms" swagger:"required"`
	EpochDurationMs       *iotajsonrpc.BigInt `json:"epoch_duration_ms" swagger:"required"`
}

func (p *Protocol) String() string {
	return string(lo.Must(json.MarshalIndent(p, "", "  ")))
}

type IotaCoinInfo struct {
	CoinType    coin.Type  `json:"coinType" swagger:"desc(BaseToken's Cointype),required"`
	Name        string     `json:"name" swagger:"desc(The base token name),required"`
	Symbol      string     `json:"tickerSymbol" swagger:"desc(The ticker symbol),required"`
	Description string     `json:"description" swagger:"desc(The token description),required"`
	IconURL     string     `json:"iconUrl" swagger:"desc(The icon URL),required"`
	Decimals    uint8      `json:"decimals" swagger:"desc(The token decimals),required"`
	TotalSupply coin.Value `json:"totalSupply" swagger:"desc(The total supply of BaseToken),required"`
}

func (c *IotaCoinInfo) String() string {
	return string(lo.Must(json.MarshalIndent(c, "", "  ")))
}

func (c *IotaCoinInfo) Bytes() []byte {
	return bcs.MustMarshal(c)
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
