package tests

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/corecontracts"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/stretchr/testify/require"
)

func (e *ChainEnv) checkCoreContracts() {
	for i := range e.Chain.AllPeers {
		b, err := e.Chain.GetStateVariable(root.Contract.Hname(), root.StateVarStateInitialized, i)
		require.NoError(e.t, err)
		require.EqualValues(e.t, []byte{0xFF}, b)

		cl := e.Chain.SCClient(governance.Contract.Hname(), nil, i)
		ret, err := cl.CallView(governance.ViewGetChainInfo.Name, nil)
		require.NoError(e.t, err)

		chid, err := codec.DecodeChainID(ret.MustGet(governance.VarChainID))
		require.NoError(e.t, err)
		require.EqualValues(e.t, e.Chain.ChainID, chid)

		aid, err := codec.DecodeAgentID(ret.MustGet(governance.VarChainOwnerID))
		require.NoError(e.t, err)
		require.EqualValues(e.t, e.Chain.OriginatorID(), aid)

		desc, err := codec.DecodeString(ret.MustGet(governance.VarDescription), "")
		require.NoError(e.t, err)
		require.EqualValues(e.t, e.Chain.Description, desc)

		records, err := e.Chain.SCClient(root.Contract.Hname(), nil, i).
			CallView(root.ViewGetContractRecords.Name, nil)
		require.NoError(e.t, err)

		contractRegistry, err := root.DecodeContractRegistry(collections.NewMapReadOnly(records, root.StateVarContractRegistry))
		require.NoError(e.t, err)
		for _, rec := range corecontracts.All {
			cr := contractRegistry[rec.Hname()]
			require.NotNil(e.t, cr, "core contract %s %+v missing", rec.Name, rec.Hname())

			require.EqualValues(e.t, rec.ProgramHash, cr.ProgramHash)
			require.EqualValues(e.t, rec.Description, cr.Description)
			require.EqualValues(e.t, rec.Name, cr.Name)
		}
	}
}

func (e *ChainEnv) checkRootsOutside() {
	for _, rec := range corecontracts.All {
		recBack, err := e.findContract(rec.Name)
		require.NoError(e.t, err)
		require.NotNil(e.t, recBack)
		require.EqualValues(e.t, rec.Name, recBack.Name)
		require.EqualValues(e.t, rec.ProgramHash, recBack.ProgramHash)
		require.EqualValues(e.t, rec.Description, recBack.Description)
	}
}

func (e *ChainEnv) getBalanceOnChain(agentID isc.AgentID, assetID []byte, nodeIndex ...int) uint64 {
	idx := 0
	if len(nodeIndex) > 0 {
		idx = nodeIndex[0]
	}
	ret, err := e.Chain.Cluster.WaspClient(idx).CallView(
		e.Chain.ChainID, accounts.Contract.Hname(), accounts.ViewBalance.Name,
		dict.Dict{
			accounts.ParamAgentID: agentID.Bytes(),
		})
	if err != nil {
		return 0
	}

	actual, err := isc.FungibleTokensFromDict(ret)
	require.NoError(e.t, err)

	if bytes.Equal(assetID, isc.BaseTokenID) {
		return actual.BaseTokens
	}

	tokenSet, err := actual.Tokens.Set()
	require.NoError(e.t, err)
	tokenID, err := isc.NativeTokenIDFromBytes(assetID)
	require.NoError(e.t, err)
	return tokenSet[tokenID].Amount.Uint64()
}

func (e *ChainEnv) checkBalanceOnChain(agentID isc.AgentID, assetID []byte, expected uint64) {
	actual := e.getBalanceOnChain(agentID, assetID)
	require.EqualValues(e.t, expected, actual)
}

func (e *ChainEnv) getAccountsOnChain() []isc.AgentID {
	r, err := e.Chain.Cluster.WaspClient(0).CallView(
		e.Chain.ChainID, accounts.Contract.Hname(), accounts.ViewAccounts.Name, nil,
	)
	require.NoError(e.t, err)

	ret := make([]isc.AgentID, 0)
	for key := range r {
		aid, err := isc.AgentIDFromBytes([]byte(key))
		require.NoError(e.t, err)

		ret = append(ret, aid)
	}
	require.NoError(e.t, err)

	return ret
}

func (e *ChainEnv) getBalancesOnChain() map[string]*isc.FungibleTokens {
	ret := make(map[string]*isc.FungibleTokens)
	acc := e.getAccountsOnChain()
	for _, agentID := range acc {
		r, err := e.Chain.Cluster.WaspClient(0).CallView(
			e.Chain.ChainID, accounts.Contract.Hname(), accounts.ViewBalance.Name,
			dict.Dict{
				accounts.ParamAgentID: agentID.Bytes(),
			},
		)
		require.NoError(e.t, err)
		ret[string(agentID.Bytes())], err = isc.FungibleTokensFromDict(r)
		require.NoError(e.t, err)
	}
	return ret
}

func (e *ChainEnv) getTotalBalance() *isc.FungibleTokens {
	r, err := e.Chain.Cluster.WaspClient(0).CallView(
		e.Chain.ChainID, accounts.Contract.Hname(), accounts.ViewTotalAssets.Name, nil,
	)
	require.NoError(e.t, err)
	ret, err := isc.FungibleTokensFromDict(r)
	require.NoError(e.t, err)
	return ret
}

func (e *ChainEnv) printAccounts(title string) {
	allBalances := e.getBalancesOnChain()
	s := fmt.Sprintf("------------------------------------- %s\n", title)
	for k, bals := range allBalances {
		aid, err := isc.AgentIDFromBytes([]byte(k))
		require.NoError(e.t, err)
		s += fmt.Sprintf("     %s\n", aid.String())
		s += fmt.Sprintf("%s\n", bals.String())
	}
	fmt.Println(s)
}

func (e *ChainEnv) checkLedger() {
	balances := e.getBalancesOnChain()
	sum := isc.NewEmptyAssets()
	for _, bal := range balances {
		sum.Add(bal)
	}
	require.True(e.t, sum.Equals(e.getTotalBalance()))
}

func (e *ChainEnv) getChainInfo() (*isc.ChainID, isc.AgentID) {
	ret, err := e.Chain.Cluster.WaspClient(0).CallView(
		e.Chain.ChainID, governance.Contract.Hname(), governance.ViewGetChainInfo.Name, nil,
	)
	require.NoError(e.t, err)

	chainID, err := codec.DecodeChainID(ret.MustGet(governance.VarChainID))
	require.NoError(e.t, err)

	ownerID, err := codec.DecodeAgentID(ret.MustGet(governance.VarChainOwnerID))
	require.NoError(e.t, err)
	return chainID, ownerID
}

func (e *ChainEnv) findContract(name string, nodeIndex ...int) (*root.ContractRecord, error) {
	i := 0
	if len(nodeIndex) > 0 {
		i = nodeIndex[0]
	}

	hname := isc.Hn(name)
	ret, err := e.Chain.Cluster.WaspClient(i).CallView(
		e.Chain.ChainID, root.Contract.Hname(), root.ViewFindContract.Name,
		dict.Dict{
			root.ParamHname: codec.EncodeHname(hname),
		})
	if err != nil {
		return nil, err
	}
	recBin, err := ret.Get(root.ParamContractRecData)
	if err != nil {
		return nil, err
	}
	return root.ContractRecordFromBytes(recBin)
}

// region waitUntilProcessed ///////////////////////////////////////////////////

const pollPeriod = 500 * time.Millisecond

func waitTrue(timeout time.Duration, fun func() bool) bool {
	deadline := time.Now().Add(timeout)
	for {
		if fun() {
			return true
		}
		time.Sleep(pollPeriod)
		if time.Now().After(deadline) {
			return false
		}
	}
}

func (e *ChainEnv) counterEquals(expected int64) conditionFn {
	return func(t *testing.T, nodeIndex int) bool {
		ret, err := e.Chain.Cluster.WaspClient(nodeIndex).CallView(
			e.Chain.ChainID, nativeIncCounterSCHname, inccounter.ViewGetCounter.Name, nil,
		)
		if err != nil {
			e.t.Logf("chainEnv::counterEquals: failed to call GetCounter: %v", err)
			return false
		}
		counter, err := codec.DecodeInt64(ret.MustGet(inccounter.VarCounter), 0)
		require.NoError(t, err)
		t.Logf("chainEnv::counterEquals: node %d: counter: %d, waiting for: %d", nodeIndex, counter, expected)
		return counter == expected
	}
}

func (e *ChainEnv) accountExists(agentID isc.AgentID) conditionFn {
	return func(t *testing.T, nodeIndex int) bool {
		return e.getBalanceOnChain(agentID, isc.BaseTokenID, nodeIndex) > 0
	}
}

func (e *ChainEnv) contractIsDeployed() conditionFn {
	return func(t *testing.T, nodeIndex int) bool {
		ret, err := e.findContract(nativeIncCounterSCName, nodeIndex)
		if err != nil {
			return false
		}
		return ret.Name == nativeIncCounterSCName
	}
}

func (e *ChainEnv) balanceOnChainBaseTokensEquals(agentID isc.AgentID, expected uint64) conditionFn {
	return func(t *testing.T, nodeIndex int) bool {
		have := e.getBalanceOnChain(agentID, isc.BaseTokenID, nodeIndex)
		e.t.Logf("chainEnv::balanceOnChainBaseTokensEquals: node=%v, have=%v, expected=%v", nodeIndex, have, expected)
		return expected == have
	}
}

type conditionFn func(t *testing.T, nodeIndex int) bool

func waitUntil(t *testing.T, fn conditionFn, nodeIndexes []int, timeout time.Duration, logMsg ...string) {
	for _, nodeIndex := range nodeIndexes {
		if len(logMsg) > 0 {
			t.Logf("-->Waiting for '%s' on node %v...", logMsg[0], nodeIndex)
		}
		w := waitTrue(timeout, func() bool {
			return fn(t, nodeIndex)
		})
		if !w {
			if len(logMsg) > 0 {
				t.Errorf("-->Waiting for %s on node %v... FAILED after %v", logMsg[0], nodeIndex, timeout)
			} else {
				t.Errorf("-->Waiting on node %v... FAILED after %v", nodeIndex, timeout)
			}
			t.Helper()
			t.Fatal()
		}
	}
}

// endregion ///////////////////////////////////////////////////////////////
