package tests

import (
	"context"
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/clients/apiextensions"
	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/corecontracts"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/core/testcore/contracts/inccounter"
	"github.com/iotaledger/wasp/packages/webapi/models"
)

func (e *ChainEnv) checkCoreContracts() {
	for i := range e.Chain.AllPeers {
		cl := e.Chain.Client(nil, i)
		ret, err := cl.CallView(context.Background(), governance.ViewGetChainInfo.Message())
		require.NoError(e.t, err)
		info, err := governance.ViewGetChainInfo.DecodeOutput(ret)
		require.NoError(e.t, err)

		require.EqualValues(e.t, e.Chain.OriginatorID(), info.ChainAdmin)

		records, err := e.Chain.Client(nil, i).
			CallView(context.Background(), root.ViewGetContractRecords.Message())
		require.NoError(e.t, err)

		contractRegistry, err := root.ViewGetContractRecords.DecodeOutput(records)
		require.NoError(e.t, err)
		for _, rec := range corecontracts.All {
			foundHname := slices.ContainsFunc(contractRegistry, func(tuple lo.Tuple2[*isc.Hname, *root.ContractRecord]) bool {
				return *tuple.A == rec.Hname()
			})
			require.True(e.t, foundHname, "core contract %s %+v missing", rec.Name, rec.Hname())
		}
	}
}

func (e *ChainEnv) checkRootsOutside() {
	for _, rec := range corecontracts.All {
		recBack, err := e.findContract(rec.Name)
		require.NoError(e.t, err)
		require.NotNil(e.t, recBack)
		require.EqualValues(e.t, rec.Name, recBack.Name)
	}
}

func (e *ChainEnv) GetL1Balance(addr *iotago.Address, coinType coin.Type) coin.Value {
	l1client := e.Chain.Cluster.L1Client()
	getBalance, err := l1client.GetBalance(context.TODO(), iotaclient.GetBalanceRequest{Owner: addr})
	require.NoError(e.t, err)
	return coin.Value(getBalance.TotalBalance.Uint64())
}

func (e *ChainEnv) GetL2Balance(agentID isc.AgentID, coinType coin.Type, nodeIndex ...int) coin.Value {
	idx := 0
	if len(nodeIndex) > 0 {
		idx = nodeIndex[0]
	}

	balance, _, err := e.Chain.Cluster.WaspClient(idx).CorecontractsAPI.
		AccountsGetAccountBalance(context.Background(), agentID.String()).
		Execute()
	require.NoError(e.t, err)

	assets, err := apiextensions.AssetsFromAPIResponse(balance)
	require.NoError(e.t, err)

	return assets.CoinBalance(coinType)
}

func (e *ChainEnv) getBalanceOnChain(agentID isc.AgentID, coinType coin.Type, nodeIndex ...int) coin.Value {
	idx := 0
	if len(nodeIndex) > 0 {
		idx = nodeIndex[0]
	}

	balance, _, err := e.Chain.Cluster.WaspClient(idx).CorecontractsAPI.
		AccountsGetAccountBalance(context.Background(), agentID.String()).
		Execute()
	require.NoError(e.t, err)

	assets, err := apiextensions.AssetsFromAPIResponse(balance)
	require.NoError(e.t, err)

	return assets.CoinBalance(coinType)
}

func (e *ChainEnv) checkBalanceOnChain(agentID isc.AgentID, coinType coin.Type, expected coin.Value) {
	actual := e.getBalanceOnChain(agentID, coinType)
	require.EqualValues(e.t, expected, actual)
}

func (e *ChainEnv) getChainInfo() (isc.ChainID, isc.AgentID) {
	chainInfo, _, err := e.Chain.Cluster.WaspClient(0).ChainsAPI.
		GetChainInfo(context.Background()).
		Execute()
	require.NoError(e.t, err)

	chainID, err := isc.ChainIDFromString(chainInfo.ChainID)
	require.NoError(e.t, err)

	admin, err := isc.AgentIDFromString(chainInfo.ChainAdmin)
	require.NoError(e.t, err)

	return chainID, admin
}

func (e *ChainEnv) findContract(name string, nodeIndex ...int) (*root.ContractRecord, error) {
	i := 0
	if len(nodeIndex) > 0 {
		i = nodeIndex[0]
	}

	hname := isc.Hn(name)

	// TODO: Validate with develop
	ret, err := apiextensions.CallView(
		context.Background(),
		e.Chain.Cluster.WaspClient(i),

		apiextensions.CallViewReq(root.ViewFindContract.Message(hname)),
	)

	require.NoError(e.t, err)

	found, recBin, err := root.ViewFindContract.DecodeOutput(ret)
	require.NoError(e.t, err)

	if !found {
		return nil, nil
	}

	return *recBin, nil
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

func (e *ChainEnv) balanceEquals(agentID isc.AgentID, amount int) conditionFn {
	return func(t *testing.T, nodeIndex int) bool {
		ret, err := apiextensions.CallView(
			context.Background(),
			e.Chain.Cluster.WaspClient(nodeIndex),
			apiclient.ContractCallViewRequest{
				ContractHName: accounts.Contract.Hname().String(),
				FunctionHName: accounts.ViewBalanceBaseToken.Hname().String(),
				Arguments:     models.ToCallArgumentsJSON(accounts.ViewBalanceBaseToken.Message(&agentID).Params),
			})
		if err != nil {
			e.t.Logf("chainEnv::counterEquals: failed to call GetCounter: %v", err)
			return false
		}

		balance, err := accounts.ViewBalanceBaseToken.DecodeOutput(ret)

		fmt.Printf("CURRENT BALANCE: %d, EXPECTED: %d\n", balance, amount)

		return coin.Value(amount) == balance
	}
}

func (e *ChainEnv) counterEquals(expected int64) conditionFn {
	return func(t *testing.T, nodeIndex int) bool {
		ret, err := apiextensions.CallView(
			context.Background(),
			e.Chain.Cluster.WaspClient(nodeIndex),
			apiclient.ContractCallViewRequest{
				ContractHName: inccounter.Contract.Hname().String(),
				FunctionHName: inccounter.ViewGetCounter.Hname().String(),
			})
		if err != nil {
			e.t.Logf("chainEnv::counterEquals: failed to call GetCounter: %v", err)
			return false
		}
		counter, err := inccounter.ViewGetCounter.DecodeOutput(ret)
		require.NoError(t, err)
		t.Logf("chainEnv::counterEquals: node %d: counter: %d, waiting for: %d", nodeIndex, counter, expected)
		return counter == expected
	}
}

func (e *ChainEnv) accountExists(agentID isc.AgentID) conditionFn {
	return func(t *testing.T, nodeIndex int) bool {
		return e.getBalanceOnChain(agentID, coin.BaseTokenType, nodeIndex) > 0
	}
}

func (e *ChainEnv) contractIsDeployed() conditionFn {
	return func(t *testing.T, nodeIndex int) bool {
		ret, err := e.findContract(inccounter.Contract.Name, nodeIndex)
		if err != nil {
			return false
		}
		return ret.Name == inccounter.Contract.Name
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

func setupNativeInccounterTest(t *testing.T, clusterSize int, committee []int, dirnameOpt ...string) *ChainEnv {
	quorum := uint16((2*len(committee))/3 + 1)

	dirname := ""
	if len(dirnameOpt) > 0 {
		dirname = dirnameOpt[0]
	}
	clu := newCluster(t, waspClusterOpts{
		nNodes:  clusterSize,
		dirName: dirname,
	})

	addr, err := clu.RunDKG(committee, quorum)
	require.NoError(t, err)

	t.Logf("generated state address: %s", addr.String())

	chain, err := clu.DeployChain(clu.Config.AllNodes(), committee, quorum, addr)
	require.NoError(t, err)
	t.Logf("deployed chainID: %s", chain.ChainID)

	e := &ChainEnv{
		t:     t,
		Clu:   clu,
		Chain: chain,
	}

	return e
}
