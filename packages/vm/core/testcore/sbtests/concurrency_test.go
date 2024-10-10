package sbtests

import (
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common/math"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/iota-go/suiclient"
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/testcore/sbtests/sbtestsc"
)

func TestCounter(t *testing.T) { run2(t, testCounter) }
func testCounter(t *testing.T) {
	_, chain := setupChain(t, nil)
	setupTestSandboxSC(t, chain, nil)

	req := solo.NewCallParamsEx(ScName, sbtestsc.FuncIncCounter.Name).AddBaseTokens(1 * isc.Million).WithGasBudget(math.MaxUint64)
	for i := 0; i < 33; i++ {
		_, err := chain.PostRequestSync(req, nil)
		require.NoError(t, err)
	}

	ret, err := chain.CallViewEx(ScName, sbtestsc.FuncGetCounter.Name)
	require.NoError(t, err)

	deco := kvdecoder.New(ret, chain.Log())
	res := deco.MustGetInt64(sbtestsc.VarCounter)
	require.EqualValues(t, 33, res)
}

func TestConcurrency(t *testing.T) { run2(t, testConcurrency) }
func testConcurrency(t *testing.T) {
	_, chain := setupChain(t, nil)
	setupTestSandboxSC(t, chain, nil)

	commonAccountInitialBalance := chain.L2BaseTokens(accounts.CommonAccount())

	req := solo.NewCallParamsEx(ScName, sbtestsc.FuncIncCounter.Name).
		AddBaseTokens(1000).WithGasBudget(math.MaxUint64)

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

	ret, err := chain.CallViewEx(ScName, sbtestsc.FuncGetCounter.Name)
	require.NoError(t, err)

	deco := kvdecoder.New(ret, chain.Log())
	res := deco.MustGetInt64(sbtestsc.VarCounter)
	require.EqualValues(t, sum, res)

	commonAccountFinalBalance := chain.L2BaseTokens(accounts.CommonAccount())
	require.Equal(t, commonAccountFinalBalance, commonAccountInitialBalance)

	contractAgentID := isc.NewContractAgentID(chain.ChainID, HScName) // SC has no funds (because it never claims funds from allowance)
	chain.AssertL2BaseTokens(contractAgentID, 0)
}

func TestConcurrency2(t *testing.T) { run2(t, testConcurrency2) }
func testConcurrency2(t *testing.T) {
	_, chain := setupChain(t, nil)
	setupTestSandboxSC(t, chain, nil)

	commonAccountInitialBalance := chain.L2BaseTokens(accounts.CommonAccount())

	var baseTokensSentPerRequest coin.Value = 1 * isc.Million
	req := solo.NewCallParamsEx(ScName, sbtestsc.FuncIncCounter.Name).
		AddBaseTokens(baseTokensSentPerRequest).WithGasBudget(math.MaxUint64)

	_, estimate, err := chain.EstimateGasOnLedger(req, nil)
	require.NoError(t, err)

	repeats := []int{300, 100, 100, 100, 200, 100, 100}
	users := make([]*cryptolib.KeyPair, len(repeats))
	userAddr := make([]*cryptolib.Address, len(repeats))
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

	ret, err := chain.CallViewEx(ScName, sbtestsc.FuncGetCounter.Name)
	require.NoError(t, err)

	deco := kvdecoder.New(ret, chain.Log())
	res := deco.MustGetInt64(sbtestsc.VarCounter)
	require.EqualValues(t, sum, res)

	for i := range users {
		expectedBalance := coin.Value(repeats[i]) * (baseTokensSentPerRequest - estimate.GasFeeCharged)
		chain.AssertL2BaseTokens(isc.NewAddressAgentID(userAddr[i]), expectedBalance)
		chain.Env.AssertL1BaseTokens(userAddr[i], suiclient.FundsFromFaucetAmount-coin.Value(repeats[i])*baseTokensSentPerRequest)
	}

	commonAccountFinalBalance := chain.L2BaseTokens(accounts.CommonAccount())
	require.Equal(t, commonAccountFinalBalance, commonAccountInitialBalance)
}
