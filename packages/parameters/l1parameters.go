package parameters

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/hive.go/log"

	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/hashing"
)

// L1Params describes parameters coming from the L1Params node
type L1Params struct {
	Protocol  *Protocol  `json:"protocol" swagger:"required"`
	BaseToken *BaseToken `json:"baseToken" swagger:"required"`
}

func (l *L1Params) String() string {
	b, err := json.MarshalIndent(l, "", "\t")
	if err != nil {
		panic(err)
	}
	return string(b)
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

func (l *L1Params) Clone() *L1Params {
	protocol := l.Protocol.Clone()
	basetoken := l.BaseToken.Clone()
	return &L1Params{
		Protocol:  &protocol,
		BaseToken: &basetoken,
	}
}

type Protocol struct {
	Epoch                 *iotajsonrpc.BigInt `json:"epoch" swagger:"required"`
	ProtocolVersion       *iotajsonrpc.BigInt `json:"protocol_version" swagger:"required"`
	SystemStateVersion    *iotajsonrpc.BigInt `json:"system_state_version" swagger:"required"`
	IotaTotalSupply       *iotajsonrpc.BigInt `json:"iota_total_supply" swagger:"required"`
	ReferenceGasPrice     *iotajsonrpc.BigInt `json:"reference_gas_price" swagger:"required"`
	EpochStartTimestampMs *iotajsonrpc.BigInt `json:"epoch_start_timestamp_ms" swagger:"required"`
	EpochDurationMs       *iotajsonrpc.BigInt `json:"epoch_duration_ms" swagger:"required"`
}

func (p *Protocol) String() string {
	b, err := json.MarshalIndent(p, "", "\t")
	if err != nil {
		panic(err)
	}
	return string(b)
}

func (p *Protocol) Clone() Protocol {
	return Protocol{
		Epoch:                 p.Epoch.Clone(),
		ProtocolVersion:       p.ProtocolVersion.Clone(),
		SystemStateVersion:    p.SystemStateVersion.Clone(),
		IotaTotalSupply:       p.IotaTotalSupply.Clone(),
		ReferenceGasPrice:     p.ReferenceGasPrice.Clone(),
		EpochStartTimestampMs: p.EpochStartTimestampMs.Clone(),
		EpochDurationMs:       p.EpochDurationMs.Clone(),
	}
}

type BaseToken struct {
	Name            string    `json:"name" swagger:"desc(The base token name),required"`
	TickerSymbol    string    `json:"tickerSymbol" swagger:"desc(The ticker symbol),required"`
	Unit            string    `json:"unit" swagger:"desc(The token unit),required"`
	Subunit         string    `json:"subunit" swagger:"desc(The token subunit),required"`
	Decimals        uint8     `json:"decimals" swagger:"desc(The token decimals),required"`
	UseMetricPrefix bool      `json:"useMetricPrefix" swagger:"desc(Whether or not the token uses a metric prefix),required"`
	CoinType        coin.Type `json:"coinType" swagger:"desc(BaseToken's Cointype),required"`
	TotalSupply     uint64    `json:"totalSupply" swagger:"desc(The total supply of BaseToken),required"`
}

func (b *BaseToken) String() string {
	ret, err := json.MarshalIndent(b, "", "\t")
	if err != nil {
		panic(err)
	}
	return string(ret)
}

func (b *BaseToken) Clone() BaseToken {
	return BaseToken{
		Name:            b.Name,
		TickerSymbol:    b.TickerSymbol,
		Unit:            b.Unit,
		Subunit:         b.Subunit,
		Decimals:        b.Decimals,
		UseMetricPrefix: b.UseMetricPrefix,
		CoinType:        b.CoinType,
		TotalSupply:     b.TotalSupply,
	}
}

const (
	Decimals    = 9
	NetworkName = "testnet"
)

var BaseTokenDefault = &BaseToken{
	Name:            "Iota",
	TickerSymbol:    "MIOTA",
	Unit:            "MIOTA",
	Subunit:         "IOTA",
	Decimals:        Decimals,
	UseMetricPrefix: false,
	CoinType:        coin.BaseTokenType,
}

// refer ProtocolConfig.max_tx_size_bytes
const MaxPayloadSize = 128 * 1024

var (
	l1ParamsMutex = &sync.RWMutex{}
	// this should set to init in the beginning, otherwise, L2 may use outdated data to calculate
	l1Params *L1Params = L1Default

	L1Default = &L1Params{
		Protocol: &Protocol{
			Epoch:                 iotajsonrpc.NewBigInt(100),
			ProtocolVersion:       iotajsonrpc.NewBigInt(1),
			SystemStateVersion:    iotajsonrpc.NewBigInt(1),
			IotaTotalSupply:       iotajsonrpc.NewBigInt(9978371123948460000),
			ReferenceGasPrice:     iotajsonrpc.NewBigInt(1000),
			EpochStartTimestampMs: iotajsonrpc.NewBigInt(1734538812318),
			EpochDurationMs:       iotajsonrpc.NewBigInt(86400000),
		},
		BaseToken: BaseTokenDefault,
	}
)

func L1() *L1Params {
	l1ParamsMutex.RLock()
	defer l1ParamsMutex.RUnlock()
	return l1Params
}

func InitL1(client iotaclient.Client, logger log.Logger) error {
	var protocol Protocol
	var totalSupply *iotajsonrpc.Supply
	timeout := 600 * time.Second
	for {
		ctx := context.Background()

		summary, err := client.GetLatestIotaSystemState(ctx)
		if err != nil {
			logger.LogError("can't get latest epoch: ", err)
			time.Sleep(60 * time.Second)
			continue
		}
		protocol = Protocol{
			Epoch:                 summary.Epoch,
			ProtocolVersion:       summary.ProtocolVersion,
			SystemStateVersion:    summary.SystemStateVersion,
			IotaTotalSupply:       summary.IotaTotalSupply,
			ReferenceGasPrice:     summary.ReferenceGasPrice,
			EpochStartTimestampMs: summary.EpochStartTimestampMs,
			EpochDurationMs:       summary.EpochDurationMs,
		}

		ctx, cancel := context.WithTimeout(ctx, timeout)
		totalSupply, err = client.GetTotalSupply(ctx, iotajsonrpc.IotaCoinType.String())
		cancel()
		if err != nil {
			logger.LogError("can't get latest total supply: ", err)
			time.Sleep(60 * time.Second)
			continue
		}
		break
	}
	if totalSupply == nil {
		return fmt.Errorf("can't get Latest L1Params")
	}

	l1ParamsMutex.Lock()
	newL1Params := l1Params.Clone()
	newL1Params.Protocol = &protocol
	newL1Params.BaseToken.TotalSupply = totalSupply.Value.Uint64()
	l1Params = newL1Params
	l1ParamsMutex.Unlock()
	return nil
}
