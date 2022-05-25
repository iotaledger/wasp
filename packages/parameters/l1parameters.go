package parameters

import (
	"os"
	"strings"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
)

// DO NOT CHANGE THIS VAR. This global var is used to get L1 protocol parameters at runtime - it must only be set by nodeconn (after obtained from L1 node)
var L1 *L1Params

func init() {
	// setup testing parameters when running in the context of tests
	if strings.HasSuffix(os.Args[0], ".test") ||
		strings.HasSuffix(os.Args[0], ".test.exe") {
		InitL1ForTesting()
	}
}

// L1Params describes parameters coming from the L1Params node
type L1Params struct {
	MaxTransactionSize int
	Protocol           *iotago.ProtocolParameters
}

func InitL1ForTesting() {
	L1 = &L1Params{
		// There are no limits on how big from a size perspective an essence can be, so it is just derived from 32KB - Message fields without payload = max size of the payload
		MaxTransactionSize: 32000,
		// https://github.com/muXxer/protocol-rfcs/blob/master/text/0032-dust-protection/0032-dust-protection.md
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
	}
}
