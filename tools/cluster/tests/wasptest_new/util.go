package wasptest

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/cbalances"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/datatypes"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
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

		contractRegistry := datatypes.NewMustMap(state, root.VarContractRegistry)

		crBytes := contractRegistry.GetAt(root.Interface.Hname().Bytes())
		require.NotNil(t, crBytes)
		rec := root.NewContractRecord(root.Interface, coretypes.AgentID{})
		require.True(t, bytes.Equal(crBytes, util.MustBytes(&rec)))

		crBytes = contractRegistry.GetAt(blob.Interface.Hname().Bytes())
		require.NotNil(t, crBytes)
		cr, err := root.DecodeContractRecord(crBytes)
		check(err, t)

		require.EqualValues(t, blob.Interface.ProgramHash, cr.ProgramHash)
		require.EqualValues(t, blob.Interface.Description, cr.Description)
		require.EqualValues(t, 0, cr.OwnerFee)
		require.EqualValues(t, blob.Interface.Name, cr.Name)

		crBytes = contractRegistry.GetAt(accounts.Interface.Hname().Bytes())
		require.NotNil(t, crBytes)
		cr, err = root.DecodeContractRecord(crBytes)
		check(err, t)

		require.EqualValues(t, accounts.Interface.ProgramHash, cr.ProgramHash)
		require.EqualValues(t, accounts.Interface.Description, cr.Description)
		require.EqualValues(t, 0, cr.OwnerFee)
		require.EqualValues(t, accounts.Interface.Name, cr.Name)
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
	require.EqualValues(t, coretypes.AgentID{}, recRoot.Creator)

	origAgentID := coretypes.NewAgentIDFromAddress(*chain.OriginatorAddress())

	recBlob, err := findContract(chain, blob.Interface.Name)
	check(err, t)
	require.NotNil(t, recBlob)
	require.EqualValues(t, blob.Interface.Name, recBlob.Name)
	require.EqualValues(t, blob.Interface.ProgramHash, recBlob.ProgramHash)
	require.EqualValues(t, blob.Interface.Description, recBlob.Description)
	require.EqualValues(t, origAgentID, recBlob.Creator)

	recAccounts, err := findContract(chain, accounts.Interface.Name)
	check(err, t)
	require.NotNil(t, recAccounts)
	require.EqualValues(t, accounts.Interface.Name, recAccounts.Name)
	require.EqualValues(t, accounts.Interface.ProgramHash, recAccounts.ProgramHash)
	require.EqualValues(t, accounts.Interface.Description, recAccounts.Description)
	require.EqualValues(t, origAgentID, recAccounts.Creator)
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
	ret, err := chain.Cluster.WaspClient(0).CallView(
		chain.ContractID(accounts.Interface.Hname()),
		accounts.FuncBalance,
		dict.FromGoMap(map[kv.Key][]byte{
			accounts.ParamAgentID: agentID[:],
		}),
	)
	check(err, t)

	actual, _, err := codec.DecodeInt64(ret.MustGet(kv.Key(color[:])))
	check(err, t)

	return actual
}

func checkBalanceOnChain(t *testing.T, chain *cluster.Chain, agentID coretypes.AgentID, color balance.Color, expected int64) {
	actual := getAgentBalanceOnChain(t, chain, agentID, color)
	require.EqualValues(t, expected, actual)
}

func getAccountsOnChain(t *testing.T, chain *cluster.Chain) []coretypes.AgentID {
	r, err := chain.Cluster.WaspClient(0).CallView(
		chain.ContractID(accounts.Interface.Hname()),
		accounts.FuncAccounts,
		nil,
	)
	check(err, t)

	ret := make([]coretypes.AgentID, 0)
	for key := range r {
		aid, err := coretypes.NewAgentIDFromBytes([]byte(key))
		check(err, t)

		ret = append(ret, aid)
	}
	check(err, t)

	return ret
}

func getBalancesOnChain(t *testing.T, chain *cluster.Chain) map[coretypes.AgentID]map[balance.Color]int64 {
	ret := make(map[coretypes.AgentID]map[balance.Color]int64)
	acc := getAccountsOnChain(t, chain)
	for _, agentID := range acc {
		r, err := chain.Cluster.WaspClient(0).CallView(
			chain.ContractID(accounts.Interface.Hname()),
			accounts.FuncBalance,
			dict.FromGoMap(map[kv.Key][]byte{
				accounts.ParamAgentID: agentID[:],
			}),
		)
		check(err, t)
		ret[agentID] = balancesDictToMap(t, r)
	}
	return ret
}

func getTotalBalance(t *testing.T, chain *cluster.Chain) map[balance.Color]int64 {
	r, err := chain.Cluster.WaspClient(0).CallView(
		chain.ContractID(accounts.Interface.Hname()),
		accounts.FuncTotalAssets,
		nil,
	)
	check(err, t)
	return balancesDictToMap(t, r)
}

func balancesDictToMap(t *testing.T, d dict.Dict) map[balance.Color]int64 {
	ret := make(map[balance.Color]int64)
	for key, value := range d {
		col, _, err := balance.ColorFromBytes([]byte(key))
		check(err, t)
		v, err := util.Int64From8Bytes(value)
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

func diffBalancesOnChain(t *testing.T, chain *cluster.Chain) coretypes.ColoredBalances {
	balances := getBalancesOnChain(t, chain)
	sum := make(map[balance.Color]int64)
	for _, bal := range balances {
		for col, b := range bal {
			s, _ := sum[col]
			sum[col] = s + b
		}
	}

	total := cbalances.NewFromMap(getTotalBalance(t, chain))
	return cbalances.NewFromMap(sum).Diff(total)
}

func checkLedger(t *testing.T, chain *cluster.Chain) {
	diff := diffBalancesOnChain(t, chain)
	if diff == nil || diff.Len() == 0 {
		return
	}
	fmt.Printf("\ninconsistent ledger %s\n", diff.String())
	require.EqualValues(t, 0, diff.Len())
}

func getChainInfo(t *testing.T, chain *cluster.Chain) (coretypes.ChainID, coretypes.AgentID) {
	ret, err := chain.Cluster.WaspClient(0).CallView(
		chain.ContractID(root.Interface.Hname()),
		root.FuncGetChainInfo,
		nil,
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
	d := dict.New()
	d.Set(root.ParamHname, codec.EncodeHname(hname))
	ret, err := chain.Cluster.WaspClient(0).CallView(
		chain.ContractID(root.Interface.Hname()),
		root.FuncFindContract,
		d,
	)
	if err != nil {
		return nil, err
	}
	recBin, err := ret.Get(root.ParamData)
	if err != nil {
		return nil, err
	}
	return root.DecodeContractRecord(recBin)
}
