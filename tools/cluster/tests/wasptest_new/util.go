package wasptest

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coret"
	"github.com/iotaledger/wasp/packages/coret/cbalances"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/accountsc"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/blob"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
	"github.com/iotaledger/wasp/tools/cluster"
	"github.com/stretchr/testify/require"
	"testing"
)

func checkRoots(t *testing.T, chain *cluster.Chain) {
	chain.WithSCState(root.Interface.Hname(), func(host string, blockIndex uint32, state codec.ImmutableMustCodec) bool {
		require.EqualValues(t, []byte{0xFF}, state.Get(root.VarStateInitialized))

		chid, _ := state.GetChainID(root.VarChainID)
		require.EqualValues(t, &chain.ChainID, chid)

		aid, _ := state.GetAgentID(root.VarChainOwnerID)
		require.EqualValues(t, *chain.OriginatorID(), *aid)

		desc, _ := state.GetString(root.VarDescription)
		require.EqualValues(t, chain.Description, desc)

		contractRegistry := state.GetMap(root.VarContractRegistry)

		crBytes := contractRegistry.GetAt(root.Interface.Hname().Bytes())
		require.NotNil(t, crBytes)
		require.True(t, bytes.Equal(crBytes, util.MustBytes(&root.RootContractRecord)))

		crBytes = contractRegistry.GetAt(blob.Interface.Hname().Bytes())
		require.NotNil(t, crBytes)
		cr, err := root.DecodeContractRecord(crBytes)
		check(err, t)

		require.EqualValues(t, blob.Interface.ProgramHash, cr.ProgramHash)
		require.EqualValues(t, blob.Interface.Description, cr.Description)
		require.EqualValues(t, 0, cr.NodeFee)
		require.EqualValues(t, blob.Interface.Name, cr.Name)

		crBytes = contractRegistry.GetAt(accountsc.Interface.Hname().Bytes())
		require.NotNil(t, crBytes)
		cr, err = root.DecodeContractRecord(crBytes)
		check(err, t)

		require.EqualValues(t, accountsc.Interface.ProgramHash, cr.ProgramHash)
		require.EqualValues(t, accountsc.Interface.Description, cr.Description)
		require.EqualValues(t, 0, cr.NodeFee)
		require.EqualValues(t, accountsc.Interface.Name, cr.Name)
		return true
	})
}

func checkRootsOutside(t *testing.T, chain *cluster.Chain) {
	recRoot, err := findContract(chain, root.Interface.Name)
	check(err, t)
	require.NotNil(t, recRoot)
	require.EqualValues(t, root.Interface.Name, recRoot.Name)
	require.EqualValues(t, root.Interface.ProgramHash, recRoot.ProgramHash)
	require.EqualValues(t, root.Interface.Description, recRoot.Description)
	require.EqualValues(t, coret.AgentID{}, recRoot.Originator)

	origAgentID := coret.NewAgentIDFromAddress(*chain.OriginatorAddress())

	recBlob, err := findContract(chain, blob.Interface.Name)
	check(err, t)
	require.NotNil(t, recBlob)
	require.EqualValues(t, blob.Interface.Name, recBlob.Name)
	require.EqualValues(t, blob.Interface.ProgramHash, recBlob.ProgramHash)
	require.EqualValues(t, blob.Interface.Description, recBlob.Description)
	require.EqualValues(t, origAgentID, recBlob.Originator)

	recAccounts, err := findContract(chain, accountsc.Interface.Name)
	check(err, t)
	require.NotNil(t, recAccounts)
	require.EqualValues(t, accountsc.Interface.Name, recAccounts.Name)
	require.EqualValues(t, accountsc.Interface.ProgramHash, recAccounts.ProgramHash)
	require.EqualValues(t, accountsc.Interface.Description, recAccounts.Description)
	require.EqualValues(t, origAgentID, recAccounts.Originator)
}

func requestFunds(wasps *cluster.Cluster, addr *address.Address, who string) error {
	err := wasps.NodeClient.RequestFunds(addr)
	if err != nil {
		return err
	}
	if !wasps.VerifyAddressBalances(addr, testutil.RequestFundsAmount, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount,
	}, "requested funds for "+who) {
		return errors.New("unexpected requested amount")
	}
	return nil
}

func getAgentBalanceOnChain(t *testing.T, chain *cluster.Chain, agentID coret.AgentID, color balance.Color) int64 {
	ret, err := chain.Cluster.WaspClient(0).CallView(
		chain.ContractID(accountsc.Interface.Hname()),
		accountsc.FuncBalance,
		dict.FromGoMap(map[kv.Key][]byte{
			accountsc.ParamAgentID: agentID[:],
		}),
	)
	check(err, t)

	c := codec.NewCodec(ret)
	actual, _, err := c.GetInt64(kv.Key(color[:]))
	check(err, t)

	return actual
}

func checkBalanceOnChain(t *testing.T, chain *cluster.Chain, agentID coret.AgentID, color balance.Color, expected int64) {
	actual := getAgentBalanceOnChain(t, chain, agentID, color)
	require.EqualValues(t, expected, actual)
}

func getAccountsOnChain(t *testing.T, chain *cluster.Chain) []coret.AgentID {
	r, err := chain.Cluster.WaspClient(0).CallView(
		chain.ContractID(accountsc.Interface.Hname()),
		accountsc.FuncAccounts,
		nil,
	)
	check(err, t)

	ret := make([]coret.AgentID, 0)
	c := codec.NewCodec(r)
	err = c.Iterate("", func(key kv.Key, value []byte) bool {
		aid, err := coret.NewAgentIDFromBytes([]byte(key))
		check(err, t)

		ret = append(ret, aid)
		return true
	})
	check(err, t)

	return ret
}

func getBalancesOnChain(t *testing.T, chain *cluster.Chain) map[coret.AgentID]map[balance.Color]int64 {
	ret := make(map[coret.AgentID]map[balance.Color]int64)
	accounts := getAccountsOnChain(t, chain)
	for _, agentID := range accounts {
		r, err := chain.Cluster.WaspClient(0).CallView(
			chain.ContractID(accountsc.Interface.Hname()),
			accountsc.FuncBalance,
			dict.FromGoMap(map[kv.Key][]byte{
				accountsc.ParamAgentID: agentID[:],
			}),
		)
		check(err, t)
		c := codec.NewCodec(r)
		ret[agentID] = make(map[balance.Color]int64)
		err = c.Iterate("", func(key kv.Key, value []byte) bool {
			col, _, err := balance.ColorFromBytes([]byte(key))
			check(err, t)
			v, err := util.Int64From8Bytes(value)
			check(err, t)
			ret[agentID][col] = v
			return true
		})
		check(err, t)
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

func diffBalancesOnChain(t *testing.T, chain *cluster.Chain) coret.ColoredBalances {
	balances := getBalancesOnChain(t, chain)
	totalAssets, ok := balances[accountsc.TotalAssetsAccountID]
	require.True(t, ok)
	sum := make(map[balance.Color]int64)
	for aid, bal := range balances {
		if aid == accountsc.TotalAssetsAccountID {
			continue
		}
		for col, b := range bal {
			s, _ := sum[col]
			sum[col] = s + b
		}
	}
	sum1 := cbalances.NewFromMap(sum)
	total := cbalances.NewFromMap(totalAssets)
	return sum1.Diff(total)
}

func checkLedger(t *testing.T, chain *cluster.Chain) {
	diff := diffBalancesOnChain(t, chain)
	if diff == nil || diff.Len() == 0 {
		return
	}
	fmt.Printf("\ninconsistent ledger %s\n", diff.String())
	require.EqualValues(t, 0, diff.Len())
}

func getChainInfo(t *testing.T, chain *cluster.Chain) (coret.ChainID, coret.AgentID) {
	ret, err := chain.Cluster.WaspClient(0).CallView(
		chain.ContractID(root.Interface.Hname()),
		root.FuncGetInfo,
		nil,
	)
	check(err, t)

	c := codec.NewCodec(ret)
	chainID, ok, err := c.GetChainID(root.VarChainID)
	check(err, t)
	require.True(t, ok)

	ownerID, ok, err := c.GetAgentID(root.VarChainOwnerID)
	check(err, t)
	require.True(t, ok)
	return *chainID, *ownerID
}

func findContract(chain *cluster.Chain, name string) (*root.ContractRecord, error) {
	hname := coret.Hn(name)
	d := dict.New()
	codec.NewCodec(d).SetHname(root.ParamHname, hname)
	ret, err := chain.Cluster.WaspClient(0).CallView(
		chain.ContractID(root.Interface.Hname()),
		root.FuncFindContract,
		d,
	)
	if err != nil {
		return nil, err
	}
	c := codec.NewCodec(ret)
	recBin, err := c.Get(root.ParamData)
	if err != nil {
		return nil, err
	}
	return root.DecodeContractRecord(recBin)
}
