package parameters

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/packages/coin"
)

// L1Params describes parameters coming from the L1Params node
type L1Params struct {
	MaxPayloadSize int        `json:"maxPayloadSize" swagger:"required"`
	Protocol       *Protocol  `json:"protocol" swagger:"required"`
	BaseToken      *BaseToken `json:"baseToken" swagger:"required"`
}

func (l *L1Params) String() string {
	ret := "{\n"
	ret += fmt.Sprintf("\tMaxPayloadSize: %d\n", l.MaxPayloadSize)
	ret += fmt.Sprintf("\tProtocol: %s\n", l.Protocol.String())
	ret += fmt.Sprintf("\tBaseToken: %s\n", l.BaseToken.String())
	ret += "}\n"
	return ret
}

func (l *L1Params) Clone() *L1Params {
	protocol := l.Protocol.Clone()
	basetoken := l.BaseToken.Clone()
	return &L1Params{
		MaxPayloadSize: l.MaxPayloadSize,
		Protocol:       &protocol,
		BaseToken:      &basetoken,
	}
}

type Protocol struct {
	Epoch                 *iotajsonrpc.BigInt
	ProtocolVersion       *iotajsonrpc.BigInt
	SystemStateVersion    *iotajsonrpc.BigInt
	IotaTotalSupply       *iotajsonrpc.BigInt
	ReferenceGasPrice     *iotajsonrpc.BigInt
	EpochStartTimestampMs *iotajsonrpc.BigInt
	EpochDurationMs       *iotajsonrpc.BigInt
}

func (p *Protocol) String() string {
	ret := "{\n"
	ret += fmt.Sprintf("\tEpoch: %s\n", p.Epoch)
	ret += fmt.Sprintf("\tProtocolVersion: %s\n", p.ProtocolVersion)
	ret += fmt.Sprintf("\tSystemStateVersion: %s\n", p.SystemStateVersion)
	ret += fmt.Sprintf("\tIotaTotalSupply: %s\n", p.IotaTotalSupply)
	ret += fmt.Sprintf("\tReferenceGasPrice: %s\n", p.ReferenceGasPrice)
	ret += fmt.Sprintf("\tEpochStartTimestampMs: %s\n", p.EpochStartTimestampMs)
	ret += fmt.Sprintf("\tEpochDurationMs: %s\n", p.EpochDurationMs)
	ret += "}\n"
	return ret
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
	ret := "{\n"
	ret += fmt.Sprintf("\tName: %s\n", b.Name)
	ret += fmt.Sprintf("\tTickerSymbol: %s\n", b.TickerSymbol)
	ret += fmt.Sprintf("\tUnit: %s\n", b.Unit)
	ret += fmt.Sprintf("\tSubunit: %s\n", b.Subunit)
	ret += fmt.Sprintf("\tDecimals: %d\n", b.Decimals)
	ret += fmt.Sprintf("\tUseMetricPrefix: %v\n", b.UseMetricPrefix)
	ret += fmt.Sprintf("\tCoinType: %s\n", b.CoinType)
	ret += fmt.Sprintf("\tTokenSupply: %d\n", b.TotalSupply)
	ret += "}\n"
	return ret
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

// for test only
func InitStaticL1(l1 *L1Params) {
	l1Params = l1
}

func InitL1(client iotaclient.Client, logger *logger.Logger) {
	var protocol Protocol
	var totalSupply *iotajsonrpc.Supply
	timeout := 600 * time.Second
	for {
		ctx := context.Background()
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		summary, err := client.GetLatestIotaSystemState(ctx)
		if err != nil {
			logger.Errorln("can't get latest epoch: ", err)
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

		totalSupply, err = client.GetTotalSupply(ctx, iotajsonrpc.IotaCoinType.String())
		if err != nil {
			logger.Errorln("can't get latest total supply: ", err)
			time.Sleep(60 * time.Second)
			continue
		}
		break
	}

	l1ParamsMutex.Lock()
	newL1Params := l1Params.Clone()
	newL1Params.Protocol = &protocol
	newL1Params.BaseToken.TotalSupply = totalSupply.Value.Uint64()
	l1Params = newL1Params
	l1ParamsMutex.Unlock()
}

type L1Syncer struct {
	client         *iotaclient.Client
	epochID        iotago.EpochId
	startTimestamp time.Time
	epochDuration  time.Duration
	timeout        time.Duration // GetLatestIotaSystemState timeout
	logger         *logger.Logger
	status         bool // switch on/off for L1Syncer
}

func NewL1Syncer(
	client *iotaclient.Client,
	timeout time.Duration,
	logger *logger.Logger,
) *L1Syncer {
	if l1Params == nil {
		l1Params = L1Default
	}
	return &L1Syncer{
		client:  client,
		timeout: timeout,
		logger:  logger,
	}
}

// Check latest L1 system parameters every hour
func (l *L1Syncer) Start() {
	l.status = true
	for l.status {
		l.Upate()
		time.Sleep(time.Until(l.startTimestamp.Add(l.epochDuration)))
	}
}

// Update the latest L1 system parameters
func (l *L1Syncer) Upate() {
	var protocol Protocol
	var newEpochID uint64
	var totalSupply *iotajsonrpc.Supply
	for {
		ctx := context.Background()
		ctx, cancel := context.WithTimeout(ctx, l.timeout)
		defer cancel()

		summary, err := l.client.GetLatestIotaSystemState(ctx)
		if err != nil {
			l.logger.Errorln("can't get latest epoch: ", err)
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

		newEpochID = summary.Epoch.Uint64()
		if newEpochID == l.epochID {
			l.logger.Debugln("same epoch: ", newEpochID)
			time.Sleep(60 * time.Second)
			continue
		}

		totalSupply, err = l.client.GetTotalSupply(ctx, iotajsonrpc.IotaCoinType.String())
		if err != nil {
			l.logger.Errorln("can't get latest total supply: ", err)
			time.Sleep(60 * time.Second)
			continue
		}
		break
	}

	l.epochID = newEpochID
	// theoretically we can check once a day only, then startTimestamp will be needed
	l.startTimestamp = l.startTimestamp.UTC()
	l.epochDuration = time.Duration(protocol.EpochDurationMs.Uint64()) * time.Millisecond

	l1ParamsMutex.Lock()
	newL1Params := l1Params.Clone()
	newL1Params.Protocol = &protocol
	newL1Params.BaseToken.TotalSupply = totalSupply.Value.Uint64()
	l1Params = newL1Params
	l1ParamsMutex.Unlock()
}

func (l *L1Syncer) Stop() {
	l.status = false
}
