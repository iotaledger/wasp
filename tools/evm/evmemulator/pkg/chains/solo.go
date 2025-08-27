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
	"github.com/iotaledger/wasp/v2/tools/evm/evmemulator/pkg/log"
)

const MaxPreFundAmount = 10_000

// hive tests require chain ID to be 1
const hiveChainID = 1

func InitSolo(genesis *core.Genesis) (*SoloContext, *solo.Chain) {
	ctx := &SoloContext{}

	env := solo.New(ctx, &solo.InitOptions{Debug: log.DebugFlag, PrintStackTrace: log.DebugFlag})

	chainAdmin, _ := env.NewKeyPairWithFunds()
	feePolicy := gas.DefaultFeePolicy()
	feePolicy.EVMGasRatio = util.Ratio32{A: 1, B: 10_00_000_000}
	chain, _ := env.NewChainExtWithGenesis(chainAdmin, 1*isc.Million, "evmemulator", hiveChainID, feePolicy, emulator.BlockKeepAll, genesis)

	// prefund the account against genesis
	for addr, acc := range genesis.Alloc {
		randDepositorSeed := []byte("GetL2FundsFromFaucet" + fmt.Sprintf("%d", rand.Int()))

		// limit the amount of the prefund
		preFundAmount := coin.Value(0)
		if acc.Balance.Uint64() > MaxPreFundAmount {
			preFundAmount = MaxPreFundAmount
		} else {
			preFundAmount = coin.Value(acc.Balance.Uint64())
		}
		acc.Balance = preFundAmount.BigInt()
		chain.GetL2FundsFromFaucetWithDepositor(isc.NewEthereumAddressAgentID(addr), randDepositorSeed, preFundAmount)
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
