package inccounter

import (
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/solo/solobench"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/vm/core/corecontracts"
	"github.com/iotaledger/wasp/packages/vm/core/root"
)

func checkCounter(e *solo.Chain, expected int64) {
	ret, err := e.CallView(ViewGetCounter.Message())
	require.NoError(e.Env.T, err)
	c, err := codec.Int64.Decode(ret.Get(VarCounter))
	require.NoError(e.Env.T, err)
	require.EqualValues(e.Env.T, expected, c)
}

func initSolo(t *testing.T) *solo.Solo {
	return solo.New(t, &solo.InitOptions{
		AutoAdjustStorageDeposit: true,
		Debug:                    true,
		PrintStackTrace:          true,
	}).WithNativeContract(Processor)
}

func TestDeployInc(t *testing.T) {
	env := initSolo(t)
	chain := env.NewChain()

	err := chain.DeployContract(nil, Contract.Name, Contract.ProgramHash)
	require.NoError(t, err)
	chain.CheckChain()
	_, _, contracts := chain.GetInfo()
	require.EqualValues(t, len(corecontracts.All)+1, len(contracts))
	checkCounter(chain, 0)
	chain.CheckAccountLedger()
}

func TestDeployIncInitParams(t *testing.T) {
	env := initSolo(t)
	chain := env.NewChain()

	err := chain.DeployContract(nil, Contract.Name, Contract.ProgramHash, InitParams(17))
	require.NoError(t, err)
	checkCounter(chain, 17)
	chain.CheckAccountLedger()
}

func TestIncDefaultParam(t *testing.T) {
	env := initSolo(t)
	chain := env.NewChain()

	err := chain.DeployContract(nil, Contract.Name, Contract.ProgramHash, InitParams(17))
	require.NoError(t, err)
	checkCounter(chain, 17)

	req := solo.NewCallParams(FuncIncCounter.Message(nil)).
		AddBaseTokens(1).
		WithMaxAffordableGasBudget()
	_, err = chain.PostRequestSync(req, nil)
	require.NoError(t, err)
	checkCounter(chain, 18)
	chain.CheckAccountLedger()
}

func TestIncParam(t *testing.T) {
	env := initSolo(t)
	chain := env.NewChain()

	err := chain.DeployContract(nil, Contract.Name, Contract.ProgramHash, InitParams(17))
	require.NoError(t, err)
	checkCounter(chain, 17)

	n := int64(3)
	req := solo.NewCallParams(FuncIncCounter.Message(&n)).
		AddBaseTokens(1).
		WithMaxAffordableGasBudget()
	_, err = chain.PostRequestSync(req, nil)
	require.NoError(t, err)
	checkCounter(chain, 20)

	chain.CheckAccountLedger()
}

func TestIncWith1Post(t *testing.T) {
	env := initSolo(t)
	chain := env.NewChain()

	err := chain.DeployContract(nil, Contract.Name, Contract.ProgramHash, InitParams(17))
	require.NoError(t, err)
	checkCounter(chain, 17)

	chain.WaitForRequestsMark()

	req := solo.NewCallParams(FuncIncAndRepeatOnceAfter2s.Message()).
		AddBaseTokens(2 * isc.Million).
		WithAllowance(isc.NewAssetsBaseTokens(1 * isc.Million)).
		WithMaxAffordableGasBudget()
	_, err = chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	// advance logical clock to unlock that timelocked request
	env.AdvanceClockBy(6 * time.Second)
	require.True(t, chain.WaitForRequestsThrough(2))

	checkCounter(chain, 19)
	chain.CheckAccountLedger()
}

func TestSpawn(t *testing.T) {
	env := initSolo(t)
	chain := env.NewChain()

	err := chain.DeployContract(nil, Contract.Name, Contract.ProgramHash, InitParams(17))
	require.NoError(t, err)
	checkCounter(chain, 17)

	nameNew := "spawnedContract"
	req := solo.NewCallParams(FuncSpawn.Message(nameNew)).
		AddBaseTokens(1).WithMaxAffordableGasBudget()
	_, err = chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	res, err := chain.CallView(root.ViewGetContractRecords.Message())
	require.NoError(t, err)
	creg := lo.Must(root.ViewGetContractRecords.Output.Decode(res))
	require.True(t, len(creg) == len(corecontracts.All)+2)
}

func initBenchmark(b *testing.B) (*solo.Chain, []*solo.CallParams) {
	// setup: deploy the inccounter contract
	log := testlogger.NewSilentLogger(b.Name(), true)
	opts := solo.DefaultInitOptions()
	opts.Log = log
	env := solo.New(b, opts).WithNativeContract(Processor)
	chain := env.NewChain()

	err := chain.DeployContract(nil, Contract.Name, Contract.ProgramHash, InitParams(0))
	require.NoError(b, err)

	// setup: prepare N requests that call FuncIncCounter
	reqs := make([]*solo.CallParams, b.N)
	for i := 0; i < b.N; i++ {
		reqs[i] = solo.NewCallParams(FuncIncCounter.Message(nil)).AddBaseTokens(1)
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
