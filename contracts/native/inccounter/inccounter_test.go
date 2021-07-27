package inccounter

import (
	"testing"
	"time"

	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/vm/core"
	"github.com/stretchr/testify/require"
)

const incName = "incTest"

func checkCounter(e *solo.Chain, expected int64) {
	ret, err := e.CallView(incName, FuncGetCounter.Name)
	require.NoError(e.Env.T, err)
	c, ok, err := codec.DecodeInt64(ret.MustGet(VarCounter))
	require.NoError(e.Env.T, err)
	require.True(e.Env.T, ok)
	require.EqualValues(e.Env.T, expected, c)
}

func TestDeployInc(t *testing.T) {
	env := solo.New(t, false, false).WithNativeContract(Processor)
	chain := env.NewChain(nil, "chain1")

	err := chain.DeployContract(nil, incName, Contract.ProgramHash)
	require.NoError(t, err)
	chain.CheckChain()
	_, _, contracts := chain.GetInfo()
	require.EqualValues(t, len(core.AllCoreContractsByHash)+1, len(contracts))
	checkCounter(chain, 0)
	chain.CheckAccountLedger()
}

func TestDeployIncInitParams(t *testing.T) {
	env := solo.New(t, false, false).WithNativeContract(Processor)
	chain := env.NewChain(nil, "chain1")

	err := chain.DeployContract(nil, incName, Contract.ProgramHash, VarCounter, 17)
	require.NoError(t, err)
	checkCounter(chain, 17)
	chain.CheckAccountLedger()
}

func TestIncDefaultParam(t *testing.T) {
	env := solo.New(t, false, false).WithNativeContract(Processor)
	chain := env.NewChain(nil, "chain1")

	err := chain.DeployContract(nil, incName, Contract.ProgramHash, VarCounter, 17)
	require.NoError(t, err)
	checkCounter(chain, 17)

	req := solo.NewCallParams(incName, FuncIncCounter.Name).WithIotas(1)
	_, err = chain.PostRequestSync(req, nil)
	require.NoError(t, err)
	checkCounter(chain, 18)
	chain.CheckAccountLedger()
}

func TestIncParam(t *testing.T) {
	env := solo.New(t, false, false).WithNativeContract(Processor)
	chain := env.NewChain(nil, "chain1")

	err := chain.DeployContract(nil, incName, Contract.ProgramHash, VarCounter, 17)
	require.NoError(t, err)
	checkCounter(chain, 17)

	req := solo.NewCallParams(incName, FuncIncCounter.Name, VarCounter, 3).WithIotas(1)
	_, err = chain.PostRequestSync(req, nil)
	require.NoError(t, err)
	checkCounter(chain, 20)

	chain.CheckAccountLedger()
}

func TestIncWith1Post(t *testing.T) {
	env := solo.New(t, false, false).WithNativeContract(Processor)
	chain := env.NewChain(nil, "chain1")

	err := chain.DeployContract(nil, incName, Contract.ProgramHash, VarCounter, 17)
	require.NoError(t, err)
	checkCounter(chain, 17)

	req := solo.NewCallParams(incName, FuncIncAndRepeatOnceAfter5s.Name).WithIotas(1)
	_, err = chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	// advance logical clock to unlock that timelocked request
	env.AdvanceClockBy(6 * time.Second)
	require.True(t, chain.WaitForRequestsThrough(4))

	checkCounter(chain, 19)
	chain.CheckAccountLedger()
}

// BenchmarkEVMStorage is a benchmark for the inccounter native contract running under solo.
// run with: go test -benchmem -cpu=1 -run=' ' -bench='Bench.*'
func BenchmarkInc(b *testing.B) {
	// setup: deploy the inccounter contract
	log := testlogger.NewSilentLogger(b.Name(), true)
	env := solo.NewWithLogger(b, log).WithNativeContract(Processor)
	chain := env.NewChain(nil, "chain1")

	err := chain.DeployContract(nil, incName, Contract.ProgramHash, VarCounter, 0)
	require.NoError(b, err)

	// setup: prepare N requests that call FuncIncCounter
	reqs := make([]*solo.CallParams, b.N)
	for i := 0; i < b.N; i++ {
		reqs[i] = solo.NewCallParams(incName, FuncIncCounter.Name).WithIotas(1)
	}

	// benchmark: send the requests.
	// TODO: benchmark multiple requests per block -- requires support from solo
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err = chain.PostRequestSync(reqs[i], nil)
		require.NoError(b, err)
	}
}
