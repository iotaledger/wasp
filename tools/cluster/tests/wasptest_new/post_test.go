package wasptest

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
	"github.com/iotaledger/wasp/packages/vm/examples/inccounter"
	"github.com/stretchr/testify/require"
	"testing"
)

func deployInccounter42(t *testing.T, name string, counter int64) coretypes.ContractID {
	hname := coretypes.Hn(name)
	description := "testing contract deployment with inccounter"
	programHash = inccounter.ProgramHash

	_, err = chain.DeployContract(name, inccounter.ProgramHashStr, description, map[string]interface{}{
		inccounter.VarCounter: counter,
		root.ParamName:        name,
	})
	check(err, t)

	chain.WithSCState(root.Interface.Hname(), func(host string, blockIndex uint32, state codec.ImmutableMustCodec) bool {
		require.EqualValues(t, 2, blockIndex)
		checkRoots(t, chain)

		contractRegistry := state.GetMap(root.VarContractRegistry)
		crBytes := contractRegistry.GetAt(hname.Bytes())
		require.NotNil(t, crBytes)
		cr, err := root.DecodeContractRecord(crBytes)
		check(err, t)

		require.EqualValues(t, programHash, cr.ProgramHash)
		require.EqualValues(t, description, cr.Description)
		require.EqualValues(t, 0, cr.NodeFee)
		require.EqualValues(t, cr.Name, name)

		return true
	})

	chain.WithSCState(hname, func(host string, blockIndex uint32, state codec.ImmutableMustCodec) bool {
		counterValue, _ := state.GetInt64(inccounter.VarCounter)
		require.EqualValues(t, 42, counterValue)
		return true
	})

	// test calling root.FuncFindContractByName view function using client
	ret, err := chain.Cluster.WaspClient(0).CallView(
		chain.ContractID(root.Interface.Hname()),
		root.FuncFindContract,
		dict.FromGoMap(map[kv.Key][]byte{
			root.ParamHname: hname.Bytes(),
		}),
	)
	check(err, t)
	recb, err := ret.Get(root.ParamData)
	check(err, t)
	rec, err := root.DecodeContractRecord(recb)
	check(err, t)
	require.EqualValues(t, description, rec.Description)

	expectCounter(t, hname, counter)
	return coretypes.NewContractID(chain.ChainID, hname)
}

func expectCounter(t *testing.T, hname coretypes.Hname, counter int64) {
	c := getCounter(t, hname)
	require.EqualValues(t, counter, c)
}

func getCounter(t *testing.T, hname coretypes.Hname) int64 {
	ret, err := chain.Cluster.WaspClient(0).CallView(
		chain.ContractID(hname),
		"getCounter",
		nil,
	)
	check(err, t)

	c := codec.NewMustCodec(ret)
	counter, _ := c.GetInt64(inccounter.VarCounter)
	check(err, t)

	return counter
}

func TestPostDeployInccounter(t *testing.T) {
	setup(t, "test_cluster")

	chain, err = clu.DeployDefaultChain()
	check(err, t)

	name := "inc"
	contractID := deployInccounter42(t, name, 42)
	t.Logf("-------------- deployed contract. Name: '%s' id: %s", name, contractID.String())
}
