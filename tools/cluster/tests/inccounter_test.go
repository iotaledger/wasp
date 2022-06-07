package tests

import (
	"testing"
	"time"

	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/corecontracts"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/tools/cluster"
	"github.com/stretchr/testify/require"
)

const (
	incName        = "inccounter"
	incDescription = "IncCounter, a PoC smart contract"
)

var incHname = iscp.Hn(incName)

const (
	varCounter    = "counter"
	varNumRepeats = "numRepeats"
	varDelay      = "delay"
)

type contractWithMessageCounterEnv struct {
	*contractEnv
	counter *cluster.MessageCounter
}

func setupWithContractAndMessageCounter(t *testing.T, nrOfRequests int) *contractWithMessageCounterEnv {
	clu := newCluster(t)

	expectations := map[string]int{
		"dismissed_committee": 0,
		"state":               3 + nrOfRequests,
		//"request_out":         3 + nrOfRequests,    // not always coming from all nodes, but from quorum only
	}

	var err error

	counter, err := clu.StartMessageCounter(expectations)
	require.NoError(t, err)
	t.Cleanup(counter.Close)

	chain, err := clu.DeployDefaultChain()
	require.NoError(t, err)

	chEnv := newChainEnv(t, clu, chain)

	cEnv := chEnv.deployWasmContract(incName, incDescription, nil)
	require.NoError(t, err)

	// deposit funds onto the contract account, so it can post a L1 request
	contractAgentID := iscp.NewContractAgentID(chEnv.Chain.ChainID, incHname)
	tx, err := chEnv.NewChainClient().Post1Request(accounts.Contract.Hname(), accounts.FuncTransferAllowanceTo.Hname(), chainclient.PostRequestParams{
		Transfer: iscp.NewTokensIotas(1_500_000),
		Args: map[kv.Key][]byte{
			accounts.ParamAgentID: codec.EncodeAgentID(contractAgentID),
		},
		Allowance: iscp.NewAllowanceIotas(1_000_000),
	})
	require.NoError(chEnv.t, err)
	_, err = chEnv.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(chEnv.Chain.ChainID, tx, 30*time.Second)
	require.NoError(chEnv.t, err)

	return &contractWithMessageCounterEnv{contractEnv: cEnv, counter: counter}
}

func (e *contractWithMessageCounterEnv) postRequest(contract, entryPoint iscp.Hname, tokens int, params map[string]interface{}) {
	transfer := iscp.NewFungibleTokens(uint64(tokens), nil)
	b := iscp.NewEmptyAssets()
	if transfer != nil {
		b = transfer
	}
	tx, err := e.NewChainClient().Post1Request(contract, entryPoint, chainclient.PostRequestParams{
		Transfer: b,
		Args:     codec.MakeDict(params),
	})
	require.NoError(e.t, err)
	_, err = e.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(e.Chain.ChainID, tx, 60*time.Second)
	require.NoError(e.t, err)
	if !e.counter.WaitUntilExpectationsMet() {
		e.t.Fatal()
	}
}

func (e *contractEnv) checkSC(numRequests int) {
	for i := range e.Chain.CommitteeNodes {
		blockIndex, err := e.Chain.BlockIndex(i)
		require.NoError(e.t, err)
		require.Greater(e.t, blockIndex, uint32(numRequests+4))

		cl := e.Chain.SCClient(governance.Contract.Hname(), nil, i)
		info, err := cl.CallView(governance.ViewGetChainInfo.Name, nil)
		require.NoError(e.t, err)

		chid, err := codec.DecodeChainID(info.MustGet(governance.VarChainID))
		require.NoError(e.t, err)
		require.EqualValues(e.t, e.Chain.ChainID, chid)

		aid, err := codec.DecodeAgentID(info.MustGet(governance.VarChainOwnerID))
		require.NoError(e.t, err)
		require.EqualValues(e.t, e.Chain.OriginatorID(), aid)

		desc, err := codec.DecodeString(info.MustGet(governance.VarDescription), "")
		require.NoError(e.t, err)
		require.EqualValues(e.t, e.Chain.Description, desc)

		recs, err := e.Chain.SCClient(root.Contract.Hname(), nil, i).CallView(root.ViewGetContractRecords.Name, nil)
		require.NoError(e.t, err)

		contractRegistry, err := root.DecodeContractRegistry(collections.NewMapReadOnly(recs, root.StateVarContractRegistry))
		require.NoError(e.t, err)
		require.EqualValues(e.t, len(corecontracts.All)+1, len(contractRegistry))

		cr := contractRegistry[incHname]
		require.EqualValues(e.t, e.programHash, cr.ProgramHash)
		require.EqualValues(e.t, incName, cr.Name)
		require.EqualValues(e.t, incDescription, cr.Description)
	}
}

func (e *ChainEnv) checkCounter(expected int) {
	for i := range e.Chain.CommitteeNodes {
		counterValue, err := e.Chain.GetCounterValue(incHname, i)
		require.NoError(e.t, err)
		require.EqualValues(e.t, expected, counterValue)
	}
}

func TestIncDeployment(t *testing.T) {
	e := setupWithContractAndMessageCounter(t, 1)

	if !e.counter.WaitUntilExpectationsMet() {
		t.Fatal()
	}
	e.checkSC(0)
	e.checkCounter(0)
}

func TestIncNothing(t *testing.T) {
	testNothing(t, 2)
}

func TestInc5xNothing(t *testing.T) {
	testNothing(t, 6)
}

func testNothing(t *testing.T, numRequests int) {
	e := setupWithContractAndMessageCounter(t, numRequests)

	entryPoint := iscp.Hn("nothing")
	for i := 0; i < numRequests; i++ {
		tx, err := e.NewChainClient().Post1Request(incHname, entryPoint)
		require.NoError(t, err)
		receipts, err := e.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(e.Chain.ChainID, tx, 30*time.Second)
		require.NoError(t, err)
		require.Equal(t, 1, len(receipts))
		require.Contains(t, receipts[0].TranslatedError, vm.ErrTargetEntryPointNotFound.MessageFormat())
	}

	if !e.counter.WaitUntilExpectationsMet() {
		t.Fatal()
	}

	e.checkSC(numRequests)
	e.checkCounter(0)
}

func TestIncIncrement(t *testing.T) {
	testIncrement(t, 1)
}

func TestInc5xIncrement(t *testing.T) {
	testIncrement(t, 5)
}

func testIncrement(t *testing.T, numRequests int) {
	e := setupWithContractAndMessageCounter(t, numRequests)

	entryPoint := iscp.Hn("increment")
	for i := 0; i < numRequests; i++ {
		tx, err := e.NewChainClient().Post1Request(incHname, entryPoint)
		require.NoError(t, err)
		_, err = e.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(e.Chain.ChainID, tx, 30*time.Second)
		require.NoError(t, err)
	}

	if !e.counter.WaitUntilExpectationsMet() {
		t.Fatal()
	}

	e.checkSC(numRequests)
	e.checkCounter(numRequests)
}

func TestIncrementWithTransfer(t *testing.T) {
	e := setupWithContractAndMessageCounter(t, 2)
	t.Fail() // TODO refactor

	entryPoint := iscp.Hn("increment")
	e.postRequest(incHname, entryPoint, 42, nil)

	// if !e.clu.AssertAddressBalances(scOwnerAddr,
	// 	iscp.NewTokensIotas(utxodb.FundsFromFaucetAmount-42)) {
	// 	t.Fatal()
	// }
	// agentID := iscp.NewAgentID(e.chain.ChainID.AsAddress(), incHname)
	// actual := e.getBalanceOnChain(agentID, iscp.IotaTokenID)
	// require.EqualValues(t, 42, actual)

	// agentID = iscp.NewAgentID(scOwnerAddr, 0)
	// actual = e.getBalanceOnChain(agentID, iscp.IotaTokenID)
	// require.EqualValues(t, 0, actual)

	e.checkCounter(1)
}

func TestIncCallIncrement1(t *testing.T) {
	e := setupWithContractAndMessageCounter(t, 2)

	entryPoint := iscp.Hn("callIncrement")
	e.postRequest(incHname, entryPoint, 1, nil)

	e.checkCounter(2)
}

func TestIncCallIncrement2Recurse5x(t *testing.T) {
	e := setupWithContractAndMessageCounter(t, 2)

	entryPoint := iscp.Hn("callIncrementRecurse5x")
	e.postRequest(incHname, entryPoint, 1_000, nil)

	e.checkCounter(6)
}

func TestIncPostIncrement(t *testing.T) {
	e := setupWithContractAndMessageCounter(t, 4) // NOTE: expectations are not used in this test, so the last parameter is meaningless

	entryPoint := iscp.Hn("postIncrement")
	e.postRequest(incHname, entryPoint, 1, nil)

	e.waitUntilCounterEquals(incHname, 2, 30*time.Second)
}

func TestIncRepeatManyIncrement(t *testing.T) {
	const numRepeats = 5
	e := setupWithContractAndMessageCounter(t, numRepeats+3) // NOTE: expectations are not used in this test, so the last parameter is meaningless

	entryPoint := iscp.Hn("repeatMany")
	e.postRequest(incHname, entryPoint, numRepeats, map[string]interface{}{
		varNumRepeats: numRepeats,
	})

	e.waitUntilCounterEquals(incHname, numRepeats+1, 30*time.Second)

	for i := range e.Chain.CommitteeNodes {
		b, err := e.Chain.GetStateVariable(incHname, varCounter, i)
		require.NoError(t, err)
		counterValue, err := codec.DecodeInt64(b, 0)
		require.NoError(t, err)
		require.EqualValues(t, numRepeats+1, counterValue)

		b, err = e.Chain.GetStateVariable(incHname, varNumRepeats, i)
		require.NoError(t, err)
		repeats, err := codec.DecodeInt64(b, 0)
		require.NoError(t, err)
		require.EqualValues(t, 0, repeats)
	}
}

func TestIncLocalStateInternalCall(t *testing.T) {
	e := setupWithContractAndMessageCounter(t, 2)
	entryPoint := iscp.Hn("localStateInternalCall")
	e.postRequest(incHname, entryPoint, 0, nil)
	e.checkCounter(2)
}

func TestIncLocalStateSandboxCall(t *testing.T) {
	e := setupWithContractAndMessageCounter(t, 2)
	entryPoint := iscp.Hn("localStateSandboxCall")
	e.postRequest(incHname, entryPoint, 0, nil)
	e.checkCounter(0)
}

func TestIncLocalStatePost(t *testing.T) {
	e := setupWithContractAndMessageCounter(t, 4)
	entryPoint := iscp.Hn("localStatePost")
	e.postRequest(incHname, entryPoint, 3, nil)
	e.checkCounter(0)
}

func TestIncViewCounter(t *testing.T) {
	e := setupWithContractAndMessageCounter(t, 2)
	entryPoint := iscp.Hn("increment")
	e.postRequest(incHname, entryPoint, 0, nil)
	e.checkCounter(1)
	ret, err := e.Chain.Cluster.WaspClient(0).CallView(
		e.Chain.ChainID, incHname, "getCounter", nil,
	)
	require.NoError(t, err)

	counter, err := codec.DecodeInt64(ret.MustGet(varCounter), 0)
	require.NoError(t, err)
	require.EqualValues(t, 1, counter)
}

func TestIncCounterDelay(t *testing.T) {
	e := setupWithContractAndMessageCounter(t, 2)
	e.postRequest(incHname, iscp.Hn("increment"), 0, nil)
	e.checkCounter(1)

	e.postRequest(incHname, iscp.Hn("incrementWithDelay"), 0, map[string]interface{}{
		varDelay: int32(5), // 5s delay
	})

	time.Sleep(3 * time.Second)
	e.checkCounter(1)
	time.Sleep(3 * time.Second)
	e.checkCounter(2)
}
