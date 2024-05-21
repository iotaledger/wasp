package tests

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/clients/apiextensions"
	"github.com/iotaledger/wasp/clients/chainclient"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/corecontracts"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/root"
)

const (
	varCounter    = "counter"
	varNumRepeats = "numRepeats"
	varDelay      = "delay"
)

type contractWithMessageCounterEnv struct {
	*contractEnv
}

func setupContract(env *ChainEnv) *contractWithMessageCounterEnv {
	env.deployNativeIncCounterSC(0)

	// deposit funds onto the contract account, so it can post a L1 request
	contractAgentID := isc.NewContractAgentID(env.Chain.ChainID, inccounter.Contract.Hname())
	tx, err := env.NewChainClient().PostRequest(accounts.FuncTransferAllowanceTo.Message(contractAgentID), chainclient.PostRequestParams{
		Transfer:  isc.NewAssetsBaseTokens(1_500_000),
		Allowance: isc.NewAssetsBaseTokens(1_000_000),
	})
	require.NoError(env.t, err)
	_, err = env.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(env.Chain.ChainID, tx, false, 30*time.Second)
	require.NoError(env.t, err)

	return &contractWithMessageCounterEnv{
		contractEnv: &contractEnv{
			ChainEnv:    env,
			programHash: inccounter.Contract.ProgramHash,
		},
	}
}

func (e *contractWithMessageCounterEnv) postRequest(contract, entryPoint isc.Hname, tokens int, params map[string]interface{}) {
	transfer := isc.NewAssets(uint64(tokens), nil)
	b := isc.NewEmptyAssets()
	if transfer != nil {
		b = transfer
	}
	tx, err := e.NewChainClient().PostRequest(isc.NewMessage(contract, entryPoint, codec.MakeDict(params)), chainclient.PostRequestParams{
		Transfer: b,
	})
	require.NoError(e.t, err)
	_, err = e.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(e.Chain.ChainID, tx, false, 60*time.Second)
	require.NoError(e.t, err)
}

func (e *contractEnv) checkSC(numRequests int) {
	for i := range e.Chain.CommitteeNodes {
		blockIndex, err := e.Chain.BlockIndex(i)
		require.NoError(e.t, err)
		require.Greater(e.t, blockIndex, uint32(numRequests+4))

		cl := e.Chain.Client(nil, i)
		r, err := cl.CallView(context.Background(), governance.ViewGetChainInfo.Message())
		require.NoError(e.t, err)
		info, err := governance.ViewGetChainInfo.Output.Decode(r)
		require.NoError(e.t, err)

		require.EqualValues(e.t, e.Chain.OriginatorID(), info.ChainOwnerID)

		recs, err := e.Chain.Client(nil, i).CallView(context.Background(), root.ViewGetContractRecords.Message())
		require.NoError(e.t, err)

		contractRegistry, err := root.ViewGetContractRecords.Output.Decode(recs)
		require.NoError(e.t, err)
		require.EqualValues(e.t, len(corecontracts.All)+1, len(contractRegistry))

		cr := contractRegistry[inccounter.Contract.Hname()]
		require.EqualValues(e.t, e.programHash, cr.ProgramHash)
		require.EqualValues(e.t, inccounter.Contract.Name, cr.Name)
	}
}

func (e *ChainEnv) checkContractCounter(expected int64) {
	for i := range e.Chain.CommitteeNodes {
		counterValue, err := e.Chain.GetCounterValue(i)
		require.NoError(e.t, err)
		require.EqualValues(e.t, expected, counterValue)
	}
}

// executed in cluster_test.go
func testInvalidEntrypoint(t *testing.T, env *ChainEnv) {
	e := setupContract(env)

	numRequests := 6
	entryPoint := isc.Hn("nothing")
	for i := 0; i < numRequests; i++ {
		tx, err := e.NewChainClient().PostRequest(isc.NewMessage(inccounter.Contract.Hname(), entryPoint))
		require.NoError(t, err)
		receipts, err := e.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(e.Chain.ChainID, tx, false, 30*time.Second)
		require.NoError(t, err)
		require.Equal(t, 1, len(receipts))
		require.Contains(t, *receipts[0].ErrorMessage, vm.ErrTargetEntryPointNotFound.MessageFormat())
	}

	e.checkSC(numRequests)
	e.checkContractCounter(0)
}

// executed in cluster_test.go
func testIncrement(t *testing.T, env *ChainEnv) {
	e := setupContract(env)

	numRequests := 5

	entryPoint := isc.Hn("increment")
	for i := 0; i < numRequests; i++ {
		tx, err := e.NewChainClient().PostRequest(isc.NewMessage(inccounter.Contract.Hname(), entryPoint))
		require.NoError(t, err)
		_, err = e.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(e.Chain.ChainID, tx, false, 30*time.Second)
		require.NoError(t, err)
	}

	e.checkSC(numRequests)
	e.checkContractCounter(int64(numRequests))
}

// executed in cluster_test.go
func testIncrementWithTransfer(t *testing.T, env *ChainEnv) {
	e := setupContract(env)

	entryPoint := isc.Hn("increment")
	e.postRequest(inccounter.Contract.Hname(), entryPoint, 42, nil)

	e.checkContractCounter(1)
}

// executed in cluster_test.go
func testIncCallIncrement1(t *testing.T, env *ChainEnv) {
	e := setupContract(env)

	entryPoint := isc.Hn("callIncrement")
	e.postRequest(inccounter.Contract.Hname(), entryPoint, 1, nil)

	e.checkContractCounter(2)
}

// executed in cluster_test.go
func testIncCallIncrement2Recurse5x(t *testing.T, env *ChainEnv) {
	e := setupContract(env)

	entryPoint := isc.Hn("callIncrementRecurse5x")
	e.postRequest(inccounter.Contract.Hname(), entryPoint, 1_000, nil)

	e.checkContractCounter(6)
}

// executed in cluster_test.go
func testIncPostIncrement(t *testing.T, env *ChainEnv) {
	e := setupContract(env)

	entryPoint := isc.Hn("postIncrement")
	e.postRequest(inccounter.Contract.Hname(), entryPoint, 1, nil)

	e.waitUntilCounterEquals(2, 30*time.Second)
}

// executed in cluster_test.go
func testIncRepeatManyIncrement(t *testing.T, env *ChainEnv) {
	const numRepeats = 5
	e := setupContract(env)

	entryPoint := isc.Hn("repeatMany")
	e.postRequest(inccounter.Contract.Hname(), entryPoint, numRepeats, map[string]interface{}{
		varNumRepeats: numRepeats,
	})

	e.waitUntilCounterEquals(numRepeats+1, 30*time.Second)

	for i := range e.Chain.CommitteeNodes {
		b, err := e.Chain.GetStateVariable(inccounter.Contract.Hname(), varCounter, i)
		require.NoError(t, err)
		counterValue, err := codec.Int64.Decode(b, 0)
		require.NoError(t, err)
		require.EqualValues(t, numRepeats+1, counterValue)

		b, err = e.Chain.GetStateVariable(inccounter.Contract.Hname(), varNumRepeats, i)
		require.NoError(t, err)
		repeats, err := codec.Int64.Decode(b, 0)
		require.NoError(t, err)
		require.EqualValues(t, 0, repeats)
	}
}

// executed in cluster_test.go
func testIncLocalStateInternalCall(t *testing.T, env *ChainEnv) {
	e := setupContract(env)
	entryPoint := isc.Hn("localStateInternalCall")
	e.postRequest(inccounter.Contract.Hname(), entryPoint, 0, nil)
	e.checkContractCounter(2)
}

// executed in cluster_test.go
func testIncLocalStateSandboxCall(t *testing.T, env *ChainEnv) {
	e := setupContract(env)
	entryPoint := isc.Hn("localStateSandboxCall")
	e.postRequest(inccounter.Contract.Hname(), entryPoint, 0, nil)
	e.checkContractCounter(0)
}

// executed in cluster_test.go
func testIncLocalStatePost(t *testing.T, env *ChainEnv) {
	e := setupContract(env)
	entryPoint := isc.Hn("localStatePost")
	e.postRequest(inccounter.Contract.Hname(), entryPoint, 3, nil)
	e.checkContractCounter(0)
}

// executed in cluster_test.go
func testIncViewCounter(t *testing.T, env *ChainEnv) {
	e := setupContract(env)
	entryPoint := isc.Hn("increment")
	e.postRequest(inccounter.Contract.Hname(), entryPoint, 0, nil)
	e.checkContractCounter(1)

	ret, err := apiextensions.CallView(
		context.Background(),
		e.Chain.Cluster.WaspClient(0),
		e.Chain.ChainID.String(),
		apiclient.ContractCallViewRequest{
			ContractHName: inccounter.Contract.Hname().String(),
			FunctionName:  "getCounter",
		})
	require.NoError(t, err)

	counter, err := codec.Int64.Decode(ret.Get(varCounter), 0)
	require.NoError(t, err)
	require.EqualValues(t, 1, counter)
}

// privtangle tests have accelerated milestones (check `startCoordinator` on `privtangle.go`)
// right now each milestone is issued each 100ms which means a "1s increase" each 100ms
// executed in cluster_test.go
func testIncCounterTimelock(t *testing.T, env *ChainEnv) {
	e := setupContract(env)
	e.postRequest(inccounter.Contract.Hname(), isc.Hn("increment"), 0, nil)
	e.checkContractCounter(1)

	e.postRequest(inccounter.Contract.Hname(), isc.Hn("incrementWithDelay"), 0, map[string]interface{}{
		varDelay: int32(50), // 50s delay()
	})

	time.Sleep(3000 * time.Millisecond) // equivalent of 30s
	e.checkContractCounter(1)
	time.Sleep(3000 * time.Millisecond) // equivalent of 30s
	e.checkContractCounter(2)
}
