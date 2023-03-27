package sbtests

import (
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common/math"
	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/testutil/utxodb"
	"github.com/iotaledger/wasp/packages/vm/core/testcore/sbtests/sbtestsc"
)

func TestCounter(t *testing.T) { run2(t, testCounter) }
func testCounter(t *testing.T, w bool) {
	_, chain := setupChain(t, nil)
	setupTestSandboxSC(t, chain, nil, w)

	req := solo.NewCallParams(ScName, sbtestsc.FuncIncCounter.Name).AddBaseTokens(1 * isc.Million).WithGasBudget(math.MaxUint64)
	for i := 0; i < 33; i++ {
		_, err := chain.PostRequestSync(req, nil)
		require.NoError(t, err)
	}

	ret, err := chain.CallView(ScName, sbtestsc.FuncGetCounter.Name)
	require.NoError(t, err)

	deco := kvdecoder.New(ret, chain.Log())
	res := deco.MustGetInt64(sbtestsc.VarCounter)
	require.EqualValues(t, 33, res)
}

func TestConcurrency(t *testing.T) { run2(t, testConcurrency) }
func testConcurrency(t *testing.T, w bool) {
	_, chain := setupChain(t, nil)
	setupTestSandboxSC(t, chain, nil, w)

	commonAccountInitialBalance := chain.L2BaseTokens(chain.CommonAccount())

	req := solo.NewCallParams(ScName, sbtestsc.FuncIncCounter.Name).
		AddBaseTokens(1000).WithGasBudget(math.MaxUint64)

	_, predictedGasFee, err := chain.EstimateGasOnLedger(req, nil, true)
	require.NoError(t, err)

	repeats := []int{300, 100, 100, 100, 200, 100, 100}
	sum := 0
	for _, i := range repeats {
		sum += i
	}

	chain.WaitForRequestsMark()
	for r, n := range repeats {
		go func(_, n int) {
			for i := 0; i < n; i++ {
				tx, _, err2 := chain.RequestFromParamsToLedger(req, nil)
				require.NoError(t, err2)
				chain.Env.EnqueueRequests(tx)
			}
		}(r, n)
	}
	require.True(t, chain.WaitForRequestsThrough(sum, 180*time.Second))

	ret, err := chain.CallView(ScName, sbtestsc.FuncGetCounter.Name)
	require.NoError(t, err)

	deco := kvdecoder.New(ret, chain.Log())
	res := deco.MustGetInt64(sbtestsc.VarCounter)
	require.EqualValues(t, sum, res)

	commonAccountFinalBalance := chain.L2BaseTokens(chain.CommonAccount())
	require.Equal(t, commonAccountFinalBalance, commonAccountInitialBalance+predictedGasFee*uint64(sum))

	contractAgentID := isc.NewContractAgentID(chain.ChainID, HScName) // SC has no funds (because it never claims funds from allowance)
	chain.AssertL2BaseTokens(contractAgentID, 0)
}

func TestConcurrency2(t *testing.T) { run2(t, testConcurrency2) }
func testConcurrency2(t *testing.T, w bool) {
	_, chain := setupChain(t, nil)
	setupTestSandboxSC(t, chain, nil, w)

	commonAccountInitialBalance := chain.L2BaseTokens(chain.CommonAccount())

	baseTokensSentPerRequest := 1 * isc.Million
	req := solo.NewCallParams(ScName, sbtestsc.FuncIncCounter.Name).
		AddBaseTokens(baseTokensSentPerRequest).WithGasBudget(math.MaxUint64)

	_, predictedGasFee, err := chain.EstimateGasOnLedger(req, nil, true)
	require.NoError(t, err)

	repeats := []int{300, 100, 100, 100, 200, 100, 100}
	users := make([]*cryptolib.KeyPair, len(repeats))
	userAddr := make([]iotago.Address, len(repeats))
	sum := 0
	for _, i := range repeats {
		sum += i
	}

	chain.WaitForRequestsMark()
	for r, n := range repeats {
		go func(r, n int) {
			users[r], userAddr[r] = chain.Env.NewKeyPairWithFunds()
			for i := 0; i < n; i++ {
				tx, _, err2 := chain.RequestFromParamsToLedger(req, users[r])
				require.NoError(t, err2)
				chain.Env.EnqueueRequests(tx)
			}
		}(r, n)
	}
	require.True(t, chain.WaitForRequestsThrough(sum, 180*time.Second))

	ret, err := chain.CallView(ScName, sbtestsc.FuncGetCounter.Name)
	require.NoError(t, err)

	deco := kvdecoder.New(ret, chain.Log())
	res := deco.MustGetInt64(sbtestsc.VarCounter)
	require.EqualValues(t, sum, res)

	for i := range users {
		expectedBalance := uint64(repeats[i]) * (baseTokensSentPerRequest - predictedGasFee)
		chain.AssertL2BaseTokens(isc.NewAgentID(userAddr[i]), expectedBalance)
		chain.Env.AssertL1BaseTokens(userAddr[i], utxodb.FundsFromFaucetAmount-uint64(repeats[i])*baseTokensSentPerRequest)
	}

	commonAccountFinalBalance := chain.L2BaseTokens(chain.CommonAccount())
	require.Equal(t, commonAccountFinalBalance, commonAccountInitialBalance+predictedGasFee*uint64(sum))
}
