package chains

import (
	"os"

	"github.com/ethereum/go-ethereum/core"

	"github.com/iotaledger/wasp/v2/packages/coin"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/solo"
	"github.com/iotaledger/wasp/v2/packages/vm/core/evm/emulator"
	"github.com/iotaledger/wasp/v2/tools/evm/evmemulator/pkg/log"
)

func InitSolo(genesis *core.Genesis) (*SoloContext, *solo.Chain) {
	ctx := &SoloContext{}

	env := solo.New(ctx, &solo.InitOptions{Debug: log.DebugFlag, PrintStackTrace: log.DebugFlag})

	chainAdmin, _ := env.NewKeyPairWithFunds()
	chain, _ := env.NewChainExtWithGenesis(chainAdmin, 1*isc.Million, "evmemulator", 1074, emulator.BlockKeepAll, genesis)

	// prefund the account against genesis
	for addr, acc := range genesis.Alloc {
		chain.GetL2FundsFromFaucet(isc.NewEthereumAddressAgentID(addr), coin.Value(acc.Balance.Uint64()))
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
