package tests

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/clients/apiextensions"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/vm/core/corecontracts"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/root"
)

func (e *ChainEnv) checkCoreContracts() {
	for i := range e.Chain.AllPeers {
		b, err := e.Chain.GetStateVariable(root.Contract.Hname(), root.StateVarStateInitialized, i)
		require.NoError(e.t, err)
		require.EqualValues(e.t, []byte{0xFF}, b)

		cl := e.Chain.SCClient(governance.Contract.Hname(), nil, i)
		ret, err := cl.CallView(context.Background(), governance.ViewGetChainInfo.Name, nil)
		require.NoError(e.t, err)

		chainID, err := codec.DecodeChainID(ret.MustGet(governance.VarChainID))
		require.NoError(e.t, err)
		require.EqualValues(e.t, e.Chain.ChainID, chainID)

		aid, err := codec.DecodeAgentID(ret.MustGet(governance.VarChainOwnerID))
		require.NoError(e.t, err)
		require.EqualValues(e.t, e.Chain.OriginatorID(), aid)

		desc, err := codec.DecodeString(ret.MustGet(governance.VarDescription), "")
		require.NoError(e.t, err)
		require.EqualValues(e.t, e.Chain.Description, desc)

		records, err := e.Chain.SCClient(root.Contract.Hname(), nil, i).
			CallView(context.Background(), root.ViewGetContractRecords.Name, nil)
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

	balance, _, err := e.Chain.Cluster.WaspClient(idx).CorecontractsApi.
		AccountsGetAccountBalance(context.Background(), e.Chain.ChainID.String(), agentID.String()).
		Execute()
	require.NoError(e.t, err)

	assets, err := apiextensions.NewAssetsFromAPIResponse(balance)
	require.NoError(e.t, err)

	if bytes.Equal(assetID, isc.BaseTokenID) {
		return assets.BaseTokens
	}

	nativeTokenID, err := isc.NativeTokenIDFromBytes(assetID)
	require.NoError(e.t, err)

	for _, nativeToken := range assets.NativeTokens {
		if nativeToken.ID.Matches(nativeTokenID) {
			// TODO: Validate bigint to uint64 behavior
			return nativeToken.Amount.Uint64()
		}
	}
	// TODO: Throw error when native token id wasn't found?
	return 0
}

func (e *ChainEnv) checkBalanceOnChain(agentID isc.AgentID, assetID []byte, expected uint64) {
	actual := e.getBalanceOnChain(agentID, assetID)
	require.EqualValues(e.t, expected, actual)
}

func (e *ChainEnv) getAccountsOnChain() []isc.AgentID {
	accounts, _, err := e.Chain.Cluster.WaspClient(0).CorecontractsApi.
		AccountsGetAccounts(context.Background(), e.Chain.ChainID.String()).
		Execute()

	require.NoError(e.t, err)

	ret := make([]isc.AgentID, 0)
	for _, address := range accounts.Accounts {
		aid, err := isc.NewAgentIDFromString(address)
		require.NoError(e.t, err)

		ret = append(ret, aid)
	}
	require.NoError(e.t, err)

	return ret
}

func (e *ChainEnv) getBalancesOnChain() map[string]*isc.Assets {
	ret := make(map[string]*isc.Assets)
	acc := e.getAccountsOnChain()

	for _, agentID := range acc {
		balance, _, err := e.Chain.Cluster.WaspClient().CorecontractsApi.
			AccountsGetAccountBalance(context.Background(), e.Chain.ChainID.String(), agentID.String()).
			Execute()
		require.NoError(e.t, err)

		assets, err := apiextensions.NewAssetsFromAPIResponse(balance)
		require.NoError(e.t, err)

		ret[string(agentID.Bytes())] = assets
	}
	return ret
}

func (e *ChainEnv) getTotalBalance() *isc.Assets {
	totalAssets, _, err := e.Chain.Cluster.WaspClient().CorecontractsApi.
		AccountsGetTotalAssets(context.Background(), e.Chain.ChainID.String()).
		Execute()
	require.NoError(e.t, err)

	assets, err := apiextensions.NewAssetsFromAPIResponse(totalAssets)
	require.NoError(e.t, err)

	return assets
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

func (e *ChainEnv) getChainInfo() (isc.ChainID, isc.AgentID) {
	chainInfo, _, err := e.Chain.Cluster.WaspClient(0).ChainsApi.
		GetChainInfo(context.Background(), e.Chain.ChainID.String()).
		Execute()
	require.NoError(e.t, err)

	chainID, err := isc.ChainIDFromString(chainInfo.ChainID)
	require.NoError(e.t, err)

	ownerID, err := isc.NewAgentIDFromString(chainInfo.ChainOwnerId)
	require.NoError(e.t, err)

	return chainID, ownerID
}

func (e *ChainEnv) findContract(name string, nodeIndex ...int) (*root.ContractRecord, error) {
	i := 0
	if len(nodeIndex) > 0 {
		i = nodeIndex[0]
	}

	hname := isc.Hn(name)

	args := dict.Dict{
		root.ParamHname: codec.EncodeHname(hname),
	}

	// TODO: Validate with develop
	ret, err := apiextensions.CallView(context.Background(), e.Chain.Cluster.WaspClient(i), apiclient.ContractCallViewRequest{
		ChainId:       e.Chain.ChainID.String(),
		ContractHName: root.Contract.Hname().String(),
		FunctionHName: root.ViewFindContract.Hname().String(),
		Arguments:     apiextensions.JSONDictToAPIJSONDict(args.JSONDict()),
	})

	require.NoError(e.t, err)

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
		ret, err := apiextensions.CallView(context.Background(), e.Chain.Cluster.WaspClient(nodeIndex), apiclient.ContractCallViewRequest{
			ChainId:       e.Chain.ChainID.String(),
			ContractHName: nativeIncCounterSCHname.String(),
			FunctionHName: inccounter.ViewGetCounter.Hname().String(),
		})
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

func setupNativeInccounterTest(t *testing.T, clusterSize int, committee []int) *ChainEnv {
	quorum := uint16((2*len(committee))/3 + 1)

	clu := newCluster(t, waspClusterOpts{nNodes: clusterSize})

	addr, err := clu.RunDKG(committee, quorum)
	require.NoError(t, err)

	t.Logf("generated state address: %s", addr.Bech32(parameters.L1().Protocol.Bech32HRP))

	chain, err := clu.DeployChain("chain", clu.Config.AllNodes(), committee, quorum, addr)
	require.NoError(t, err)
	t.Logf("deployed chainID: %s", chain.ChainID)

	e := &ChainEnv{
		t:     t,
		Clu:   clu,
		Chain: chain,
	}
	e.deployNativeIncCounterSC(0)
	return e
}
