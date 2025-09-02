package chains

import (
	"fmt"
	"math/rand"
	"os"

	"github.com/ethereum/go-ethereum/core"

	"github.com/iotaledger/wasp/v2/packages/coin"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/solo"
	"github.com/iotaledger/wasp/v2/packages/util"
	"github.com/iotaledger/wasp/v2/packages/vm/core/evm/emulator"
	"github.com/iotaledger/wasp/v2/packages/vm/gas"
	"github.com/iotaledger/wasp/v2/tools/evm/evmemulator/pkg/cli"
	"github.com/iotaledger/wasp/v2/tools/evm/evmemulator/pkg/log"
)

// hive tests require chain ID to be 1
const hiveChainID = 1
const defaultChainID = 1074

func InitSolo(genesis *core.Genesis) (*SoloContext, *solo.Chain) {
	ctx := &SoloContext{}

	env := solo.New(ctx, &solo.InitOptions{Debug: log.DebugFlag, PrintStackTrace: log.DebugFlag})

	chainAdmin, _ := env.NewKeyPairWithFunds()

	var chain *solo.Chain
	if cli.IsHive {
		// Hive: cheaper gas and chain ID = 1
		feePolicy := gas.DefaultFeePolicy()
		feePolicy.EVMGasRatio = util.Ratio32{A: 1, B: 10_00_000_000}
		chain, _ = env.NewChainExt(chainAdmin, 1*isc.Million, "evmemulator", hiveChainID, emulator.BlockKeepAll, feePolicy, genesis)

		// Prefund the accounts defined in genesis via faucet (Hive behavior)
		for addr, acc := range genesis.Alloc {
			randDepositorSeed := []byte("GetL2FundsFromFaucet" + fmt.Sprintf("%d", rand.Int()))
			chain.GetL2FundsFromFaucetWithDepositor(isc.NewEthereumAddressAgentID(addr), randDepositorSeed, coin.Value(acc.Balance.Uint64()))
		}
	} else {
		// Non-Hive: — default gas policy and default chain ID
		chain, _ = env.NewChainExt(chainAdmin, 1*isc.Million, "evmemulator", defaultChainID, emulator.BlockKeepAll, nil, nil)
		// No additional prefunding loop for genesis accounts in non-hive mode
	}

	return ctx, chain
}

type SoloContext struct {
	cleanup []func()
}

func (s *SoloContext) CleanupAll() {
	for i := len(s.cleanup) - 1; i >= 0; i-- {
		s.cleanup[i]()
	}
}

func (s *SoloContext) Cleanup(f func()) {
	s.cleanup = append(s.cleanup, f)
}

func (*SoloContext) Errorf(format string, args ...interface{}) {
	log.Printf("error: "+format, args)
}

func (*SoloContext) FailNow() {
	os.Exit(1)
}

func (s *SoloContext) Fatalf(format string, args ...any) {
	log.Printf("fatal: "+format, args)
	s.FailNow()
}

func (*SoloContext) Helper() {
}

func (*SoloContext) Logf(format string, args ...any) {
	log.Printf(format, args...)
}

func (*SoloContext) Name() string {
	return "evmemulator"
}
