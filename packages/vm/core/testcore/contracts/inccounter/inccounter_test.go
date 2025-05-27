package inccounter_test

import (
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/solo/solobench"
	"github.com/iotaledger/wasp/packages/testutil/l1starter"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/vm/core/testcore/contracts/inccounter"
)

func TestMain(m *testing.M) {
	l1starter.TestMain(m)
}

func checkCounter(e *solo.Chain, expected int64) {
	ret, err := e.CallView(inccounter.ViewGetCounter.Message())
	require.NoError(e.Env.T, err)

	output := lo.Must(inccounter.ViewGetCounter.DecodeOutput(ret))

	require.EqualValues(e.Env.T, expected, output)
}

func initSolo(t *testing.T) *solo.Solo {
	return solo.New(t, &solo.InitOptions{
		Debug:           true,
		PrintStackTrace: true,
	})
}

func TestDeployIncInitParams(t *testing.T) {
	env := initSolo(t)
	chain := env.NewChain()

	checkCounter(chain, 0)
	chain.CheckAccountLedger()
}

func TestIncDefaultParam(t *testing.T) {
	env := initSolo(t)
	chain := env.NewChain()

	checkCounter(chain, 0)

	req := solo.NewCallParams(inccounter.FuncIncCounter.Message(nil)).
		AddBaseTokens(1).
		WithMaxAffordableGasBudget()
	_, err := chain.PostRequestSync(req, nil)
	require.NoError(t, err)
	checkCounter(chain, 1)
	chain.CheckAccountLedger()
}

func TestIncParam(t *testing.T) {
	env := initSolo(t)
	chain := env.NewChain()

	checkCounter(chain, 0)

	n := int64(3)
	req := solo.NewCallParams(inccounter.FuncIncCounter.Message(&n)).
		AddBaseTokens(1).
		WithMaxAffordableGasBudget()
	_, err := chain.PostRequestSync(req, nil)
	require.NoError(t, err)
	checkCounter(chain, 3)

	chain.CheckAccountLedger()
}

func initBenchmark(b *testing.B) (*solo.Chain, []*solo.CallParams) {
	// setup: deploy the inccounter contract
	log := testlogger.NewSilentLogger(b.Name(), true)
	opts := solo.DefaultInitOptions()
	opts.Log = log
	env := solo.New(b, opts)
	chain := env.NewChain()

	// setup: prepare N requests that call FuncIncCounter
	reqs := make([]*solo.CallParams, b.N)
	for i := 0; i < b.N; i++ {
		reqs[i] = solo.NewCallParams(inccounter.FuncIncCounter.Message(nil)).AddBaseTokens(1)
	}

	return chain, reqs
}

// BenchmarkIncSync is a benchmark for the inccounter native contract running under solo,
// processing requests synchronously, and producing 1 block per request.
// run with: go test -benchmem -cpu=1 -run=' ' -bench='Bench.*'
func BenchmarkIncSync(b *testing.B) {
	chain, reqs := initBenchmark(b)
	solobench.RunBenchmarkSync(b, chain, reqs, nil)
}

// BenchmarkIncAsync is a benchmark for the inccounter native contract running under solo,
// processing requests synchronously, and producing 1 block per many requests.
// run with: go test -benchmem -cpu=1 -run=' ' -bench='Bench.*'
func BenchmarkIncAsync(b *testing.B) {
	chain, reqs := initBenchmark(b)
	solobench.RunBenchmarkAsync(b, chain, reqs, nil)
}
