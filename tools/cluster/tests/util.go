package tests

import (
	"errors"
	"fmt"
	"testing"

	"github.com/iotaledger/wasp/packages/coretypes/chainid"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/tools/cluster"
	"github.com/stretchr/testify/require"
)

func checkRoots(t *testing.T, chain *cluster.Chain) {
	chain.WithSCState(root.Interface.Hname(), func(host string, blockIndex uint32, state dict.Dict) bool {

		require.EqualValues(t, []byte{0xFF}, state.MustGet(root.VarStateInitialized))

		chid, _, _ := codec.DecodeChainID(state.MustGet(root.VarChainID))
		require.EqualValues(t, chain.ChainID, chid)

		aid, _, _ := codec.DecodeAgentID(state.MustGet(root.VarChainOwnerID))
		require.EqualValues(t, *chain.OriginatorID(), aid)

		desc, _, _ := codec.DecodeString(state.MustGet(root.VarDescription))
		require.EqualValues(t, chain.Description, desc)

		contractRegistry := collections.NewMapReadOnly(state, root.VarContractRegistry)

		for _, rec := range core.AllCoreContractsByHash {
			crBytes := contractRegistry.MustGetAt(rec.Hname().Bytes())
			require.NotNil(t, crBytes)
			cr, err := root.DecodeContractRecord(crBytes)
			check(err, t)

			require.EqualValues(t, rec.ProgramHash, cr.ProgramHash)
			require.EqualValues(t, rec.Description, cr.Description)
			require.EqualValues(t, 0, cr.OwnerFee)
			require.EqualValues(t, rec.Name, cr.Name)
		}
		//
		//crBytes := contractRegistry.MustGetAt(root.Interface.Hname().Bytes())
		//require.NotNil(t, crBytes)
		//rec := root.NewContractRecord(root.Interface, &coretypes.AgentID{})
		//require.True(t, bytes.Equal(crBytes, util.MustBytes(rec)))
		//
		//crBytes = contractRegistry.MustGetAt(blob.Interface.Hname().Bytes())
		//require.NotNil(t, crBytes)
		//cr, err := root.DecodeContractRecord(crBytes)
		//check(err, t)
		//
		//require.EqualValues(t, blob.Interface.ProgramHash, cr.ProgramHash)
		//require.EqualValues(t, blob.Interface.Description, cr.Description)
		//require.EqualValues(t, 0, cr.OwnerFee)
		//require.EqualValues(t, blob.Interface.Name, cr.Name)
		//
		//crBytes = contractRegistry.MustGetAt(accounts.Interface.Hname().Bytes())
		//require.NotNil(t, crBytes)
		//cr, err = root.DecodeContractRecord(crBytes)
		//check(err, t)
		//
		//require.EqualValues(t, accounts.Interface.ProgramHash, cr.ProgramHash)
		//require.EqualValues(t, accounts.Interface.Description, cr.Description)
		//require.EqualValues(t, 0, cr.OwnerFee)
		//require.EqualValues(t, accounts.Interface.Name, cr.Name)

		return true
	})
}

func checkRootsOutside(t *testing.T, chain *cluster.Chain) {
	for _, rec := range core.AllCoreContractsByHash {
		recBack, err := findContract(chain, rec.Name)
		check(err, t)
		require.NotNil(t, recBack)
		require.EqualValues(t, rec.Name, recBack.Name)
		require.EqualValues(t, rec.ProgramHash, recBack.ProgramHash)
		require.EqualValues(t, rec.Description, recBack.Description)
		require.True(t, recBack.Creator.IsNil())
	}
	//
	//recRoot, err := findContract(chain, root.Interface.Name)
	//check(err, t)
	//require.NotNil(t, recRoot)
	//require.EqualValues(t, root.Interface.Name, recRoot.Name)
	//require.EqualValues(t, root.Interface.ProgramHash, recRoot.ProgramHash)
	//require.EqualValues(t, root.Interface.Description, recRoot.Description)
	//require.True(t, recRoot.Creator.IsNil())
	//
	//recBlob, err := findContract(chain, blob.Interface.Name)
	//check(err, t)
	//require.NotNil(t, recBlob)
	//require.EqualValues(t, blob.Interface.Name, recBlob.Name)
	//require.EqualValues(t, blob.Interface.ProgramHash, recBlob.ProgramHash)
	//require.EqualValues(t, blob.Interface.Description, recBlob.Description)
	//require.True(t, recBlob.Creator.IsNil())
	//
	//recAccounts, err := findContract(chain, accounts.Interface.Name)
	//check(err, t)
	//require.NotNil(t, recAccounts)
	//require.EqualValues(t, accounts.Interface.Name, recAccounts.Name)
	//require.EqualValues(t, accounts.Interface.ProgramHash, recAccounts.ProgramHash)
	//require.EqualValues(t, accounts.Interface.Description, recAccounts.Description)
	//require.True(t, recAccounts.Creator.IsNil())
	//
	//recEventlog, err := findContract(chain, eventlog.Interface.Name)
	//check(err, t)
	//require.NotNil(t, recEventlog)
	//require.EqualValues(t, eventlog.Interface.Name, recEventlog.Name)
	//require.EqualValues(t, eventlog.Interface.ProgramHash, recEventlog.ProgramHash)
	//require.EqualValues(t, eventlog.Interface.Description, recEventlog.Description)
	//require.True(t, recEventlog.Creator.IsNil())
	//
	//recDefault, err := findContract(chain, _default.Interface.Name)
	//check(err, t)
	//require.NotNil(t, recDefault)
	//require.EqualValues(t, _default.Interface.Name, recDefault.Name)
	//require.EqualValues(t, _default.Interface.ProgramHash, recDefault.ProgramHash)
	//require.EqualValues(t, _default.Interface.Description, recDefault.Description)
	//require.True(t, recDefault.Creator.IsNil())
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

func getBalanceOnChain(t *testing.T, chain *cluster.Chain, agentID *coretypes.AgentID, color ledgerstate.Color) uint64 {
	ret, err := chain.Cluster.WaspClient(0).CallView(
		chain.ChainID, accounts.Interface.Hname(), accounts.FuncBalance,
		dict.Dict{
			accounts.ParamAgentID: agentID.Bytes(),
		})
	check(err, t)

	actual, _, err := codec.DecodeUint64(ret.MustGet(kv.Key(color[:])))
	check(err, t)

	return actual
}

func checkBalanceOnChain(t *testing.T, chain *cluster.Chain, agentID *coretypes.AgentID, color ledgerstate.Color, expected uint64) {
	actual := getBalanceOnChain(t, chain, agentID, color)
	require.EqualValues(t, int64(expected), int64(actual))
}

func getAccountsOnChain(t *testing.T, chain *cluster.Chain) []*coretypes.AgentID {
	r, err := chain.Cluster.WaspClient(0).CallView(
		chain.ChainID, accounts.Interface.Hname(), accounts.FuncAccounts,
	)
	check(err, t)

	ret := make([]*coretypes.AgentID, 0)
	for key := range r {
		aid, err := coretypes.NewAgentIDFromBytes([]byte(key))
		check(err, t)

		ret = append(ret, aid)
	}
	check(err, t)

	return ret
}

func getBalancesOnChain(t *testing.T, chain *cluster.Chain) map[*coretypes.AgentID]map[ledgerstate.Color]uint64 {
	ret := make(map[*coretypes.AgentID]map[ledgerstate.Color]uint64)
	acc := getAccountsOnChain(t, chain)
	for _, agentID := range acc {
		r, err := chain.Cluster.WaspClient(0).CallView(
			chain.ChainID, accounts.Interface.Hname(), accounts.FuncBalance,
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
		chain.ChainID, accounts.Interface.Hname(), accounts.FuncTotalAssets,
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
			s, _ := sum[col]
			sum[col] = s + b
		}
	}

	total := ledgerstate.NewColoredBalances(getTotalBalance(t, chain))

	require.EqualValues(t, sum, total.Map())
}

func getChainInfo(t *testing.T, chain *cluster.Chain) (chainid.ChainID, coretypes.AgentID) {
	ret, err := chain.Cluster.WaspClient(0).CallView(
		chain.ChainID, root.Interface.Hname(), root.FuncGetChainInfo,
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

func findContract(chain *cluster.Chain, name string) (*root.ContractRecord, error) {
	hname := coretypes.Hn(name)
	ret, err := chain.Cluster.WaspClient(0).CallView(
		chain.ChainID, root.Interface.Hname(), root.FuncFindContract,
		dict.Dict{
			root.ParamHname: codec.EncodeHname(hname),
		})
	if err != nil {
		return nil, err
	}
	recBin, err := ret.Get(root.ParamData)
	if err != nil {
		return nil, err
	}
	return root.DecodeContractRecord(recBin)
}
