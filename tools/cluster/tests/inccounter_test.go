package tests

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/corecontracts"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/root"
)

const (
	incName        = "inccounter"
	incDescription = "IncCounter, a PoC smart contract"
)

var incHname = isc.Hn(incName)

const (
	varCounter    = "counter"
	varNumRepeats = "numRepeats"
	varDelay      = "delay"
)

type contractWithMessageCounterEnv struct {
	*contractEnv
}

func setupWithContract(t *testing.T) *contractWithMessageCounterEnv {
	clu := newCluster(t)

	chain, err := clu.DeployDefaultChain()
	require.NoError(t, err)

	chEnv := newChainEnv(t, clu, chain)
	cEnv := chEnv.deployWasmContract(incName, incDescription, nil)

	// deposit funds onto the contract account, so it can post a L1 request
	contractAgentID := isc.NewContractAgentID(chEnv.Chain.ChainID, incHname)
	tx, err := chEnv.NewChainClient().Post1Request(accounts.Contract.Hname(), accounts.FuncTransferAllowanceTo.Hname(), chainclient.PostRequestParams{
		Transfer: isc.NewFungibleBaseTokens(1_500_000),
		Args: map[kv.Key][]byte{
			accounts.ParamAgentID: codec.EncodeAgentID(contractAgentID),
		},
		Allowance: isc.NewAllowanceBaseTokens(1_000_000),
	})
	require.NoError(chEnv.t, err)
	_, err = chEnv.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(chEnv.Chain.ChainID, tx, 30*time.Second)
	require.NoError(chEnv.t, err)

	return &contractWithMessageCounterEnv{contractEnv: cEnv}
}

func (e *contractWithMessageCounterEnv) postRequest(contract, entryPoint isc.Hname, tokens int, params map[string]interface{}) {
	transfer := isc.NewFungibleTokens(uint64(tokens), nil)
	b := isc.NewEmptyFungibleTokens()
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
}

func (e *contractEnv) checkSC(numRequests int) {
	for i := range e.Chain.CommitteeNodes {
		blockIndex, err := e.Chain.BlockIndex(i)
		require.NoError(e.t, err)
		require.Greater(e.t, blockIndex, uint32(numRequests+4))

		cl := e.Chain.SCClient(governance.Contract.Hname(), nil, i)
		info, err := cl.CallView(governance.ViewGetChainInfo.Name, nil)
		require.NoError(e.t, err)

		chainID, err := codec.DecodeChainID(info.MustGet(governance.VarChainID))
		require.NoError(e.t, err)
		require.EqualValues(e.t, e.Chain.ChainID, chainID)

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

func (e *ChainEnv) checkWasmContractCounter(expected int) {
	for i := range e.Chain.CommitteeNodes {
		counterValue, err := e.Chain.GetCounterValue(incHname, i)
		require.NoError(e.t, err)
		require.EqualValues(e.t, expected, counterValue)
	}
}

func TestIncDeployment(t *testing.T) {
	e := setupWithContract(t)

	e.checkSC(0)
	e.checkWasmContractCounter(0)
}

func TestIncNothing(t *testing.T) {
	testNothing(t, 2)
}

func TestInc5xNothing(t *testing.T) {
	testNothing(t, 6)
}

func testNothing(t *testing.T, numRequests int) {
	e := setupWithContract(t)

	entryPoint := isc.Hn("nothing")
	for i := 0; i < numRequests; i++ {
		tx, err := e.NewChainClient().Post1Request(incHname, entryPoint)
		require.NoError(t, err)
		receipts, err := e.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(e.Chain.ChainID, tx, 30*time.Second)
		require.NoError(t, err)
		require.Equal(t, 1, len(receipts))
		require.Contains(t, receipts[0].ResolvedError, vm.ErrTargetEntryPointNotFound.MessageFormat())
	}

	e.checkSC(numRequests)
	e.checkWasmContractCounter(0)
}

func TestIncIncrement(t *testing.T) {
	testIncrement(t, 1)
}

func TestInc5xIncrement(t *testing.T) {
	testIncrement(t, 5)
}

func testIncrement(t *testing.T, numRequests int) {
	e := setupWithContract(t)

	entryPoint := isc.Hn("increment")
	for i := 0; i < numRequests; i++ {
		tx, err := e.NewChainClient().Post1Request(incHname, entryPoint)
		require.NoError(t, err)
		_, err = e.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(e.Chain.ChainID, tx, 30*time.Second)
		require.NoError(t, err)
	}

	e.checkSC(numRequests)
	e.checkWasmContractCounter(numRequests)
}

func TestIncrementWithTransfer(t *testing.T) {
	e := setupWithContract(t)

	entryPoint := isc.Hn("increment")
	e.postRequest(incHname, entryPoint, 42, nil)

	e.checkWasmContractCounter(1)
}

func TestIncCallIncrement1(t *testing.T) {
	e := setupWithContract(t)

	entryPoint := isc.Hn("callIncrement")
	e.postRequest(incHname, entryPoint, 1, nil)

	e.checkWasmContractCounter(2)
}

func TestIncCallIncrement2Recurse5x(t *testing.T) {
	e := setupWithContract(t)

	entryPoint := isc.Hn("callIncrementRecurse5x")
	e.postRequest(incHname, entryPoint, 1_000, nil)

	e.checkWasmContractCounter(6)
}

func TestIncPostIncrement(t *testing.T) {
	e := setupWithContract(t)

	entryPoint := isc.Hn("postIncrement")
	e.postRequest(incHname, entryPoint, 1, nil)

	e.waitUntilCounterEquals(incHname, 2, 30*time.Second)
}

func TestIncRepeatManyIncrement(t *testing.T) {
	const numRepeats = 5
	e := setupWithContract(t)

	entryPoint := isc.Hn("repeatMany")
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
	e := setupWithContract(t)
	entryPoint := isc.Hn("localStateInternalCall")
	e.postRequest(incHname, entryPoint, 0, nil)
	e.checkWasmContractCounter(2)
}

func TestIncLocalStateSandboxCall(t *testing.T) {
	e := setupWithContract(t)
	entryPoint := isc.Hn("localStateSandboxCall")
	e.postRequest(incHname, entryPoint, 0, nil)
	e.checkWasmContractCounter(0)
}

func TestIncLocalStatePost(t *testing.T) {
	e := setupWithContract(t)
	entryPoint := isc.Hn("localStatePost")
	e.postRequest(incHname, entryPoint, 3, nil)
	e.checkWasmContractCounter(0)
}

func TestIncViewCounter(t *testing.T) {
	e := setupWithContract(t)
	entryPoint := isc.Hn("increment")
	e.postRequest(incHname, entryPoint, 0, nil)
	e.checkWasmContractCounter(1)
	ret, err := e.Chain.Cluster.WaspClient(0).CallView(
		e.Chain.ChainID, incHname, "getCounter", nil,
	)
	require.NoError(t, err)

	counter, err := codec.DecodeInt64(ret.MustGet(varCounter), 0)
	require.NoError(t, err)
	require.EqualValues(t, 1, counter)
}

// privtangle tests have accellerate milestones (check `startCoordinator` on `privtangle.go`)
// right now each milestone is issued each 100ms which means a "1s increase" each 100ms
func TestIncCounterTimelock(t *testing.T) {
	e := setupWithContract(t)
	e.postRequest(incHname, isc.Hn("increment"), 0, nil)
	e.checkWasmContractCounter(1)

	e.postRequest(incHname, isc.Hn("incrementWithDelay"), 0, map[string]interface{}{
		varDelay: int32(5), // 5s delay()
	})

	time.Sleep(300 * time.Millisecond) // equivalent of 3s
	e.checkWasmContractCounter(1)
	time.Sleep(300 * time.Millisecond) // equivalent of 3s
	e.checkWasmContractCounter(2)
}
