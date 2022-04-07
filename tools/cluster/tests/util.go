package tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/iotaledger/wasp/packages/utxodb"
	"github.com/iotaledger/wasp/packages/vm/core/corecontracts"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/stretchr/testify/require"
)

func (e *chainEnv) checkCoreContracts() {
	for i := range e.chain.CommitteeNodes {
		b, err := e.chain.GetStateVariable(root.Contract.Hname(), root.StateVarStateInitialized, i)
		require.NoError(e.t, err)
		require.EqualValues(e.t, []byte{0xFF}, b)

		cl := e.chain.SCClient(governance.Contract.Hname(), nil, i)
		ret, err := cl.CallView(governance.FuncGetChainInfo.Name, nil)
		require.NoError(e.t, err)

		chid, err := codec.DecodeChainID(ret.MustGet(governance.VarChainID))
		require.NoError(e.t, err)
		require.EqualValues(e.t, e.chain.ChainID, chid)

		aid, err := codec.DecodeAgentID(ret.MustGet(governance.VarChainOwnerID))
		require.NoError(e.t, err)
		require.EqualValues(e.t, e.chain.OriginatorID(), aid)

		desc, err := codec.DecodeString(ret.MustGet(governance.VarDescription), "")
		require.NoError(e.t, err)
		require.EqualValues(e.t, e.chain.Description, desc)

		records, err := e.chain.SCClient(root.Contract.Hname(), nil, i).
			CallView(root.FuncGetContractRecords.Name, nil)
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

func (e *chainEnv) checkRootsOutside() {
	for _, rec := range corecontracts.All {
		recBack, err := e.findContract(rec.Name)
		require.NoError(e.t, err)
		require.NotNil(e.t, recBack)
		require.EqualValues(e.t, rec.Name, recBack.Name)
		require.EqualValues(e.t, rec.ProgramHash, recBack.ProgramHash)
		require.EqualValues(e.t, rec.Description, recBack.Description)
		require.True(e.t, recBack.Creator.IsNil())
	}
}

func (e *env) requestFunds(addr iotago.Address, who string) {
	err := e.clu.RequestFunds(addr)
	require.NoError(e.t, err)
	if !e.clu.AssertAddressBalances(addr, iscp.NewTokensIotas(utxodb.FundsFromFaucetAmount)) {
		e.t.Logf("unexpected requested amount")
		e.t.FailNow()
	}
}

func (e *chainEnv) getBalanceOnChain(agentID *iscp.AgentID, assetID []byte, nodeIndex ...int) uint64 {
	idx := 0
	if len(nodeIndex) > 0 {
		idx = nodeIndex[0]
	}
	ret, err := e.chain.Cluster.WaspClient(idx).CallView(
		e.chain.ChainID, accounts.Contract.Hname(), accounts.FuncViewBalance.Name,
		dict.Dict{
			accounts.ParamAgentID: agentID.Bytes(),
		})
	if err != nil {
		return 0
	}

	actual, err := codec.DecodeUint64(ret.MustGet(kv.Key(assetID)), 0)
	require.NoError(e.t, err)

	return actual
}

func (e *chainEnv) checkBalanceOnChain(agentID *iscp.AgentID, assetID []byte, expected uint64) {
	actual := e.getBalanceOnChain(agentID, assetID)
	require.EqualValues(e.t, int64(expected), int64(actual))
}

func (e *chainEnv) getAccountsOnChain() []*iscp.AgentID {
	r, err := e.chain.Cluster.WaspClient(0).CallView(
		e.chain.ChainID, accounts.Contract.Hname(), accounts.FuncViewAccounts.Name, nil,
	)
	require.NoError(e.t, err)

	ret := make([]*iscp.AgentID, 0)
	for key := range r {
		aid, err := iscp.AgentIDFromBytes([]byte(key))
		require.NoError(e.t, err)

		ret = append(ret, aid)
	}
	require.NoError(e.t, err)

	return ret
}

func (e *chainEnv) getBalancesOnChain() map[*iscp.AgentID]*iscp.FungibleTokens {
	ret := make(map[*iscp.AgentID]*iscp.FungibleTokens)
	acc := e.getAccountsOnChain()
	for _, agentID := range acc {
		r, err := e.chain.Cluster.WaspClient(0).CallView(
			e.chain.ChainID, accounts.Contract.Hname(), accounts.FuncViewBalance.Name,
			dict.Dict{
				accounts.ParamAgentID: agentID.Bytes(),
			},
		)
		require.NoError(e.t, err)
		ret[agentID], err = iscp.FungibleTokensFromDict(r)
		require.NoError(e.t, err)
	}
	return ret
}

func (e *chainEnv) getTotalBalance() *iscp.FungibleTokens {
	r, err := e.chain.Cluster.WaspClient(0).CallView(
		e.chain.ChainID, accounts.Contract.Hname(), accounts.FuncViewTotalAssets.Name, nil,
	)
	require.NoError(e.t, err)
	ret, err := iscp.FungibleTokensFromDict(r)
	require.NoError(e.t, err)
	return ret
}

func (e *chainEnv) printAccounts(title string) {
	allBalances := e.getBalancesOnChain()
	s := fmt.Sprintf("------------------------------------- %s\n", title)
	for aid, bals := range allBalances {
		s += fmt.Sprintf("     %s\n", aid.String(e.clu.GetL1NetworkPrefix()))
		s += fmt.Sprintf("%s\n", bals.String())
	}
	fmt.Println(s)
}

func (e *chainEnv) checkLedger() {
	balances := e.getBalancesOnChain()
	sum := iscp.NewEmptyAssets()
	for _, bal := range balances {
		sum.Add(bal)
	}
	require.True(e.t, sum.Equals(e.getTotalBalance()))
}

func (e *chainEnv) getChainInfo() (*iscp.ChainID, *iscp.AgentID) {
	ret, err := e.chain.Cluster.WaspClient(0).CallView(
		e.chain.ChainID, governance.Contract.Hname(), governance.FuncGetChainInfo.Name, nil,
	)
	require.NoError(e.t, err)

	chainID, err := codec.DecodeChainID(ret.MustGet(governance.VarChainID))
	require.NoError(e.t, err)

	ownerID, err := codec.DecodeAgentID(ret.MustGet(governance.VarChainOwnerID))
	require.NoError(e.t, err)
	return chainID, ownerID
}

func (e *chainEnv) findContract(name string, nodeIndex ...int) (*root.ContractRecord, error) {
	i := 0
	if len(nodeIndex) > 0 {
		i = nodeIndex[0]
	}

	hname := iscp.Hn(name)
	ret, err := e.chain.Cluster.WaspClient(i).CallView(
		e.chain.ChainID, root.Contract.Hname(), root.FuncFindContract.Name,
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

func (e *chainEnv) counterEquals(expected int64) conditionFn {
	return func(t *testing.T, nodeIndex int) bool {
		ret, err := e.chain.Cluster.WaspClient(nodeIndex).CallView(
			e.chain.ChainID, incCounterSCHname, inccounter.FuncGetCounter.Name, nil,
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

func (e *chainEnv) accountExists(agentID *iscp.AgentID) conditionFn {
	return func(t *testing.T, nodeIndex int) bool {
		return e.getBalanceOnChain(agentID, iscp.IotaTokenID, nodeIndex) > 0
	}
}

//nolint:unparam
func (e *chainEnv) contractIsDeployed(contractName string) conditionFn {
	return func(t *testing.T, nodeIndex int) bool {
		ret, err := e.findContract(contractName, nodeIndex)
		if err != nil {
			return false
		}
		return ret.Name == contractName
	}
}

func (e *chainEnv) balanceOnChainIotaEquals(agentID *iscp.AgentID, iotas uint64) conditionFn {
	return func(t *testing.T, nodeIndex int) bool {
		have := e.getBalanceOnChain(agentID, iscp.IotaTokenID, nodeIndex)
		e.t.Logf("chainEnv::balanceOnChainIotaEquals: node=%v, have=%v, expected=%v", nodeIndex, have, iotas)
		return iotas == have
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
			t.FailNow()
		}
	}
}

// endregion ///////////////////////////////////////////////////////////////
