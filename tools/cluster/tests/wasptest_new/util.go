package wasptest

import (
	"bytes"
	"errors"
	"fmt"
	cbalances "github.com/iotaledger/wasp/packages/vm/cbalances"
	"testing"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/builtinvm"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/accountsc"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
	builtinutil "github.com/iotaledger/wasp/packages/vm/builtinvm/util"
	"github.com/iotaledger/wasp/tools/cluster"
	"github.com/stretchr/testify/require"
)

func checkRoots(t *testing.T, chain *cluster.Chain) {
	chain.WithSCState(root.Hname, func(host string, blockIndex uint32, state codec.ImmutableMustCodec) bool {
		require.EqualValues(t, []byte{0xFF}, state.Get(root.VarStateInitialized))

		chid, _ := state.GetChainID(root.VarChainID)
		require.EqualValues(t, &chain.ChainID, chid)

		aid, _ := state.GetAgentID(root.VarChainOwnerID)
		require.EqualValues(t, *chain.OriginatorID(), *aid)

		desc, _ := state.GetString(root.VarDescription)
		require.EqualValues(t, chain.Description, desc)

		contractRegistry := state.GetMap(root.VarContractRegistry)

		crBytes := contractRegistry.GetAt(root.Hname.Bytes())
		require.NotNil(t, crBytes)
		require.True(t, bytes.Equal(crBytes, util.MustBytes(&root.RootContractRecord)))

		crBytes = contractRegistry.GetAt(accountsc.Hname.Bytes())
		require.NotNil(t, crBytes)
		cr, err := root.DecodeContractRecord(crBytes)
		check(err, t)

		require.EqualValues(t, builtinvm.VMType, cr.VMType)
		require.EqualValues(t, accountsc.Description, cr.Description)
		require.EqualValues(t, 0, cr.NodeFee)
		require.EqualValues(t, builtinutil.BuiltinFullName(accountsc.Name, accountsc.Version), cr.Name)
		return true
	})
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

func getAgentBalanceOnChain(t *testing.T, chain *cluster.Chain, agentID coretypes.AgentID, color balance.Color) int64 {
	ret, err := chain.Cluster.WaspClient(0).StateView(
		chain.ContractID(accountsc.Hname),
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

func checkBalanceOnChain(t *testing.T, chain *cluster.Chain, agentID coretypes.AgentID, color balance.Color, expected int64) {
	actual := getAgentBalanceOnChain(t, chain, agentID, color)
	require.EqualValues(t, expected, actual)
}

func getAccountsOnChain(t *testing.T, chain *cluster.Chain) []coretypes.AgentID {
	r, err := chain.Cluster.WaspClient(0).StateView(
		chain.ContractID(accountsc.Hname),
		accountsc.FuncAccounts,
		nil,
	)
	check(err, t)

	ret := make([]coretypes.AgentID, 0)
	c := codec.NewCodec(r)
	err = c.Iterate("", func(key kv.Key, value []byte) bool {
		aid, err := coretypes.NewAgentIDFromBytes([]byte(key))
		check(err, t)

		ret = append(ret, aid)
		return true
	})
	check(err, t)

	return ret
}

func getBalancesOnChain(t *testing.T, chain *cluster.Chain) map[coretypes.AgentID]map[balance.Color]int64 {
	ret := make(map[coretypes.AgentID]map[balance.Color]int64)
	accounts := getAccountsOnChain(t, chain)
	for _, agentID := range accounts {
		r, err := chain.Cluster.WaspClient(0).StateView(
			chain.ContractID(accountsc.Hname),
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

func diffBalancesOnChain(t *testing.T, chain *cluster.Chain) coretypes.ColoredBalances {
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
