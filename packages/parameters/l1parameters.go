package parameters

import (
	"os"
	"strings"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
)

// L1Params describes parameters coming from the L1Params node
type L1Params struct {
	MaxTransactionSize int
	Protocol           *iotago.ProtocolParameters
	BaseToken          *BaseToken
}

type BaseToken struct {
	Name            string
	TickerSymbol    string
	Unit            string
	Subunit         string
	Decimals        uint32
	UseMetricPrefix bool
}

var (
	l1Params *L1Params

	l1ForTesting = &L1Params{
		// There are no limits on how big from a size perspective an essence can be, so it is just derived from 32KB - Message fields without payload = max size of the payload
		MaxTransactionSize: 32000,
		Protocol: &iotago.ProtocolParameters{
			Version:     tpkg.TestProtoParas.Version,
			NetworkName: tpkg.TestProtoParas.NetworkName,
			Bech32HRP:   tpkg.TestProtoParas.Bech32HRP,
			MinPoWScore: tpkg.TestProtoParas.MinPoWScore,
			RentStructure: iotago.RentStructure{
				VByteCost:    10,
				VBFactorData: 1,
				VBFactorKey:  1,
			},
			TokenSupply: tpkg.TestProtoParas.TokenSupply,
		},
		BaseToken: &BaseToken{
			Name:            "Iota",
			TickerSymbol:    "MIOTA",
			Unit:            "MIOTA",
			Subunit:         "IOTA",
			Decimals:        6,
			UseMetricPrefix: false,
		},
	}

	l1ParamsLazyInit func()
)

func isTestContext() bool {
	return strings.HasSuffix(os.Args[0], ".test") ||
		strings.HasSuffix(os.Args[0], ".test.exe") ||
		strings.HasSuffix(os.Args[0], "__debug_bin")
}

func L1() *L1Params {
	if l1Params == nil {
		if isTestContext() {
			l1Params = l1ForTesting
		} else if l1ParamsLazyInit != nil {
			l1ParamsLazyInit()
		}
	}
	if l1Params == nil {
		panic("call InitL1() first")
	}
	return l1Params
}

// InitL1Lazy sets a function to be called the first time L1() is called.
// The function must call InitL1().
func InitL1Lazy(f func()) {
	l1ParamsLazyInit = f
}

func InitL1(l1 *L1Params) {
	l1Params = l1
}
