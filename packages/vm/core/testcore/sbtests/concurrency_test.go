package sbtests

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/v2/packages/coin"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/solo"
	"github.com/iotaledger/wasp/v2/packages/vm/core/testcore/sbtests/sbtestsc"
)

func TestCounter(t *testing.T) {
	_, chain := setupChain(t)
	setupTestSandboxSC(t, chain, nil)

	req := solo.NewCallParamsEx(ScName, sbtestsc.FuncIncCounter.Name).AddBaseTokens(1 * isc.Million).WithGasBudget(math.MaxUint64)
	for i := 0; i < 33; i++ {
		_, err := chain.PostRequestSync(req, nil)
		require.NoError(t, err)
	}

	res, err := sbtestsc.FuncGetCounter.Call(func(msg isc.Message) (isc.CallArguments, error) {
		return chain.CallViewWithContract(ScName, msg)
	})
	require.NoError(t, err)
	require.EqualValues(t, 33, res)
}

func TestManyRequests(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}
	_, chain := setupChain(t)
	setupTestSandboxSC(t, chain, nil)

	gasCoinValueBefore := chain.GetLatestGasCoin().Value

	req := solo.NewCallParamsEx(ScName, sbtestsc.FuncIncCounter.Name).
		AddBaseTokens(1000).WithGasBudget(math.MaxUint64)

	const N = 100
	for range N {
		_, _, err2 := chain.SendRequest(req, nil)
		require.NoError(t, err2)
	}

	const maxRequestsPerBlock = 10
	runs := chain.RunAllReceivedRequests(maxRequestsPerBlock)
	require.EqualValues(t, N/maxRequestsPerBlock, runs)

	ret, err := chain.CallViewEx(ScName, sbtestsc.FuncGetCounter.Name)
	require.NoError(t, err)
	counterResult, err := sbtestsc.FuncGetCounter.DecodeOutput(ret)
	require.NoError(t, err)
	require.EqualValues(t, N, counterResult)

	gasCoinValueAfter := chain.GetLatestGasCoin().Value
	require.Greater(t, gasCoinValueAfter, gasCoinValueBefore)

	contractAgentID := isc.NewContractAgentID(HScName) // SC has no funds (because it never claims funds from allowance)
	chain.AssertL2BaseTokens(contractAgentID, 0)
}

func TestManyRequests2(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}
	_, chain := setupChain(t)
	setupTestSandboxSC(t, chain, nil)

	gasCoinValueBefore := chain.GetLatestGasCoin().Value

	var baseTokensSentPerRequest coin.Value = 1 * isc.Million
	req := solo.NewCallParamsEx(ScName, sbtestsc.FuncIncCounter.Name).
		AddBaseTokens(baseTokensSentPerRequest).WithGasBudget(math.MaxUint64)

	_, estimate, err := chain.EstimateGasOnLedger(req, nil)
	require.NoError(t, err)

	repeats := []int{30, 10, 10, 10, 20, 10, 10}
	users := make([]*cryptolib.KeyPair, len(repeats))
	userAddr := make([]*cryptolib.Address, len(repeats))
	l1Gas := make([]coin.Value, len(repeats))

	sum := 0
	for r, n := range repeats {
		users[r], userAddr[r] = chain.Env.NewKeyPairWithFunds()
		for range n {
			_, l1Res, err2 := chain.SendRequest(req, users[r])
			require.NoError(t, err2)
			sum++
			l1Gas[r] += coin.Value(l1Res.Effects.Data.GasFee())
		}
	}

	time.Sleep(1 * time.Second)

	const maxRequestsPerBlock = 50
	runs := chain.RunAllReceivedRequests(maxRequestsPerBlock)
	require.EqualValues(t, sum/maxRequestsPerBlock, runs)

	ret, err := chain.CallViewEx(ScName, sbtestsc.FuncGetCounter.Name)
	require.NoError(t, err)
	counterResult, err := sbtestsc.FuncGetCounter.DecodeOutput(ret)
	require.NoError(t, err)
	require.EqualValues(t, sum, counterResult)

	for i := range users {
		expectedBalance := coin.Value(repeats[i]) * (baseTokensSentPerRequest - estimate.GasFeeCharged)
		chain.AssertL2BaseTokens(isc.NewAddressAgentID(userAddr[i]), expectedBalance)
		chain.Env.AssertL1BaseTokens(userAddr[i], iotaclient.FundsFromFaucetAmount-coin.Value(repeats[i])*baseTokensSentPerRequest-l1Gas[i])
	}

	gasCoinValueAfter := chain.GetLatestGasCoin().Value
	require.Greater(t, gasCoinValueAfter, gasCoinValueBefore)
}
