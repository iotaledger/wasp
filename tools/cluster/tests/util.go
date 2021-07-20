package tests

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/webapi/model"
	"github.com/iotaledger/wasp/tools/cluster"
	"github.com/stretchr/testify/require"
)

func checkCoreContracts(t *testing.T, chain *cluster.Chain) {
	for i := range chain.CommitteeNodes {
		b, err := chain.GetStateVariable(root.Contract.Hname(), root.VarStateInitialized, i)
		require.NoError(t, err)
		require.EqualValues(t, []byte{0xFF}, b)

		cl := chain.SCClient(root.Contract.Hname(), nil, i)
		ret, err := cl.CallView(root.FuncGetChainInfo.Name, nil)
		require.NoError(t, err)

		chid, _, _ := codec.DecodeChainID(ret.MustGet(root.VarChainID))
		require.EqualValues(t, chain.ChainID, chid)

		aid, _, _ := codec.DecodeAgentID(ret.MustGet(root.VarChainOwnerID))
		require.EqualValues(t, *chain.OriginatorID(), aid)

		desc, _, _ := codec.DecodeString(ret.MustGet(root.VarDescription))
		require.EqualValues(t, chain.Description, desc)

		contractRegistry, err := root.DecodeContractRegistry(collections.NewMapReadOnly(ret, root.VarContractRegistry))
		require.NoError(t, err)
		for _, rec := range core.AllCoreContractsByHash {
			cr := contractRegistry[rec.Contract.Hname()]
			require.NotNil(t, cr, "core contract %s %+v missing", rec.Contract.Name, rec.Contract.Hname())

			require.EqualValues(t, rec.Contract.ProgramHash, cr.ProgramHash)
			require.EqualValues(t, rec.Contract.Description, cr.Description)
			require.EqualValues(t, 0, cr.OwnerFee)
			require.EqualValues(t, rec.Contract.Name, cr.Name)
		}
	}
}

func checkRootsOutside(t *testing.T, chain *cluster.Chain) {
	for _, rec := range core.AllCoreContractsByHash {
		recBack, err := findContract(chain, rec.Contract.Name)
		check(err, t)
		require.NotNil(t, recBack)
		require.EqualValues(t, rec.Contract.Name, recBack.Name)
		require.EqualValues(t, rec.Contract.ProgramHash, recBack.ProgramHash)
		require.EqualValues(t, rec.Contract.Description, recBack.Description)
		require.True(t, recBack.Creator.IsNil())
	}
}

func requestFunds(wasps *cluster.Cluster, addr ledgerstate.Address, who string) error {
	err := wasps.GoshimmerClient().RequestFunds(addr)
	if err != nil {
		return err
	}
	if !wasps.VerifyAddressBalances(addr, solo.Saldo, map[ledgerstate.Color]uint64{
		ledgerstate.ColorIOTA: solo.Saldo,
	}, "requested funds for "+who) {
		return errors.New("unexpected requested amount")
	}
	return nil
}

func getBalanceOnChain(t *testing.T, chain *cluster.Chain, agentID *iscp.AgentID, color ledgerstate.Color, nodeIndex ...int) uint64 {
	idx := 0
	if len(nodeIndex) > 0 {
		idx = nodeIndex[0]
	}
	ret, err := chain.Cluster.WaspClient(idx).CallView(
		chain.ChainID, accounts.Contract.Hname(), accounts.FuncViewBalance.Name,
		dict.Dict{
			accounts.ParamAgentID: agentID.Bytes(),
		})
	if err != nil {
		return 0
	}

	actual, _, err := codec.DecodeUint64(ret.MustGet(kv.Key(color[:])))
	check(err, t)

	return actual
}

func checkBalanceOnChain(t *testing.T, chain *cluster.Chain, agentID *iscp.AgentID, color ledgerstate.Color, expected uint64) {
	actual := getBalanceOnChain(t, chain, agentID, color)
	require.EqualValues(t, int64(expected), int64(actual))
}

func getAccountsOnChain(t *testing.T, chain *cluster.Chain) []*iscp.AgentID {
	r, err := chain.Cluster.WaspClient(0).CallView(
		chain.ChainID, accounts.Contract.Hname(), accounts.FuncViewAccounts.Name, nil,
	)
	check(err, t)

	ret := make([]*iscp.AgentID, 0)
	for key := range r {
		aid, err := iscp.NewAgentIDFromBytes([]byte(key))
		check(err, t)

		ret = append(ret, aid)
	}
	check(err, t)

	return ret
}

func getBalancesOnChain(t *testing.T, chain *cluster.Chain) map[*iscp.AgentID]map[ledgerstate.Color]uint64 {
	ret := make(map[*iscp.AgentID]map[ledgerstate.Color]uint64)
	acc := getAccountsOnChain(t, chain)
	for _, agentID := range acc {
		r, err := chain.Cluster.WaspClient(0).CallView(
			chain.ChainID, accounts.Contract.Hname(), accounts.FuncViewBalance.Name,
			dict.Dict{
				accounts.ParamAgentID: agentID.Bytes(),
			})
		check(err, t)
		ret[agentID] = balancesDictToMap(t, r)
	}
	return ret
}

func getTotalBalance(t *testing.T, chain *cluster.Chain) map[ledgerstate.Color]uint64 {
	r, err := chain.Cluster.WaspClient(0).CallView(
		chain.ChainID, accounts.Contract.Hname(), accounts.FuncViewTotalAssets.Name, nil,
	)
	check(err, t)
	return balancesDictToMap(t, r)
}

func balancesDictToMap(t *testing.T, d dict.Dict) map[ledgerstate.Color]uint64 {
	ret := make(map[ledgerstate.Color]uint64)
	for key, value := range d {
		col, _, err := ledgerstate.ColorFromBytes([]byte(key))
		check(err, t)
		v, err := util.Uint64From8Bytes(value)
		check(err, t)
		ret[col] = v
	}
	return ret
}

func printAccounts(t *testing.T, chain *cluster.Chain, title string) {
	allBalances := getBalancesOnChain(t, chain)
	s := fmt.Sprintf("------------------------------------- %s\n", title)
	for aid, bals := range allBalances {
		s += fmt.Sprintf("     %s\n", aid.String())
		for k, v := range bals {
			s += fmt.Sprintf("                %s: %d\n", k.String(), v)
		}
	}
	fmt.Println(s)
}

func checkLedger(t *testing.T, chain *cluster.Chain) {
	balances := getBalancesOnChain(t, chain)
	sum := make(map[ledgerstate.Color]uint64)
	for _, bal := range balances {
		for col, b := range bal {
			s := sum[col]
			sum[col] = s + b
		}
	}

	total := ledgerstate.NewColoredBalances(getTotalBalance(t, chain))

	require.EqualValues(t, sum, total.Map())
}

func getChainInfo(t *testing.T, chain *cluster.Chain) (iscp.ChainID, iscp.AgentID) {
	ret, err := chain.Cluster.WaspClient(0).CallView(
		chain.ChainID, root.Contract.Hname(), root.FuncGetChainInfo.Name, nil,
	)
	check(err, t)

	chainID, ok, err := codec.DecodeChainID(ret.MustGet(root.VarChainID))
	check(err, t)
	require.True(t, ok)

	ownerID, ok, err := codec.DecodeAgentID(ret.MustGet(root.VarChainOwnerID))
	check(err, t)
	require.True(t, ok)
	return chainID, ownerID
}

func findContract(chain *cluster.Chain, name string, nodeIndex ...int) (*root.ContractRecord, error) {
	i := 0
	if len(nodeIndex) > 0 {
		i = nodeIndex[0]
	}

	hname := iscp.Hn(name)
	ret, err := chain.Cluster.WaspClient(i).CallView(
		chain.ChainID, root.Contract.Hname(), root.FuncFindContract.Name,
		dict.Dict{
			root.ParamHname: codec.EncodeHname(hname),
		})
	if err != nil {
		return nil, err
	}
	recBin, err := ret.Get(root.VarData)
	if err != nil {
		return nil, err
	}
	return root.DecodeContractRecord(recBin)
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

func getHTTPError(err error) *model.HTTPError {
	if err == nil {
		return nil
	}
	httpError, ok := err.(*model.HTTPError)
	if ok {
		return httpError
	}
	return getHTTPError(errors.Unwrap(err))
}

func counterEquals(chain *cluster.Chain, expected int64) conditionFn {
	return func(t *testing.T, nodeIndex int) bool {
		ret, err := chain.Cluster.WaspClient(nodeIndex).CallView(
			chain.ChainID, incCounterSCHname, inccounter.FuncGetCounter.Name, nil,
		)
		if err != nil {
			return false
		}
		counter, _, err := codec.DecodeInt64(ret.MustGet(inccounter.VarCounter))
		require.NoError(t, err)
		t.Logf("node %d: counter: %d, waiting for: %d", nodeIndex, counter, expected)
		return counter == expected
	}
}

func contractIsDeployed(chain *cluster.Chain, contractName string) conditionFn {
	return func(t *testing.T, nodeIndex int) bool {
		ret, err := findContract(chain, contractName, nodeIndex)
		if err != nil {
			return false
		}
		return ret.Name == contractName
	}
}

func balanceOnChainIotaEquals(chain *cluster.Chain, agentID *iscp.AgentID, iotas uint64) conditionFn {
	return func(t *testing.T, nodeIndex int) bool {
		return iotas == getBalanceOnChain(t, chain, agentID, ledgerstate.ColorIOTA, nodeIndex)
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
