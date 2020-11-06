package wasptest

import (
	"bytes"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/examples"
	"github.com/iotaledger/wasp/packages/vm/examples/inccounter"
	"testing"

	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
	"github.com/stretchr/testify/require"
)

func TestDeployChain(t *testing.T) {
	clu := setup(t, "test_cluster")

	err := clu.ListenToMessages(map[string]int{
		"chainrec":            2,
		"active_committee":    1,
		"dismissed_committee": 0,
		"state":               2,
		"request_in":          1,
		"request_out":         2,
	})
	check(err, t)

	chain, err := clu.DeployDefaultChain()
	check(err, t)

	if !clu.WaitUntilExpectationsMet() {
		t.Fail()
	}

	chain.WithSCState(0, func(host string, blockIndex uint32, state codec.ImmutableMustCodec) bool {
		t.Logf("Verifying state of SC 0, node %s", host)

		require.EqualValues(t, 1, blockIndex)

		require.EqualValues(t, []byte{0xFF}, state.Get(root.VarStateInitialized))

		chid, _ := state.GetChainID(root.VarChainID)
		require.EqualValues(t, &chain.ChainID, chid)

		desc, _ := state.GetString(root.VarDescription)
		require.EqualValues(t, chain.Description, desc)

		contractRegistry := state.GetArray(root.VarContractRegistry)
		require.EqualValues(t, 1, contractRegistry.Len())

		crBytes := contractRegistry.GetAt(0)
		require.True(t, bytes.Equal(crBytes, util.MustBytes(root.GetRootContractRecord())))
		return true
	})
}

func TestDeployContract(t *testing.T) {
	clu := setup(t, "test_cluster")

	err := clu.ListenToMessages(map[string]int{
		"chainrec":            2,
		"active_committee":    1,
		"dismissed_committee": 0,
		"state":               2,
		"request_in":          1,
		"request_out":         2,
	})
	check(err, t)

	chain, err := clu.DeployDefaultChain()
	check(err, t)

	description := "testing contract deployment with inccounter"

	_, err = chain.DeployBuiltinContract(examples.VMType, inccounter.ProgramHash, description, map[string]interface{}{
		inccounter.VarCounter: 42,
	})
	check(err, t)

	if !clu.WaitUntilExpectationsMet() {
		t.Fail()
	}

	chain.WithSCState(0, func(host string, blockIndex uint32, state codec.ImmutableMustCodec) bool {
		t.Logf("Verifying state of SC 0, node %s blockIndex %d", host, blockIndex)

		require.EqualValues(t, 2, blockIndex)

		contractRegistry := state.GetArray(root.VarContractRegistry)
		require.EqualValues(t, 2, contractRegistry.Len())

		crBytes := contractRegistry.GetAt(1)
		cr, err := root.DecodeContractRecord(crBytes)
		check(err, t)

		require.EqualValues(t, examples.VMType, cr.VMType)
		require.EqualValues(t, description, cr.Description)
		require.EqualValues(t, 0, cr.NodeFee)

		return true
	})
	chain.WithSCState(1, func(host string, blockIndex uint32, state codec.ImmutableMustCodec) bool {
		t.Logf("Verifying state of SC 1, node %s blockIndex %d", host, blockIndex)

		counterValue, _ := state.GetInt64(inccounter.VarCounter)
		require.EqualValues(t, 42, counterValue)

		return true

	})
}

//
//func TestActivate1Chain(t *testing.T) {
//	clu := setup(t, "test_cluster", "TestActivate1Chain")
//
//	err := clu.ListenToMessages(map[string]int{
//		"chainrec":           2,
//		"active_committee":    1,
//		"dismissed_committee": 0,
//		"request_in":          0,
//		"request_out":         0,
//		"state":               0,
//	})
//	check(err, t)
//
//	sc := &clu.SmartContractConfig[0]
//
//	_, err = PutChainRecord(clu, sc)
//	check(err, t)
//
//	err = Activate1Chain(clu, sc)
//	check(err, t)
//
//	if !clu.WaitUntilExpectationsMet() {
//		t.Fail()
//	}
//}
//
//func TestActivateAllChains(t *testing.T) {
//	clu := setup(t, "test_cluster", "TestActivateAllSC")
//
//	err := clu.ListenToMessages(map[string]int{
//		"chainrec":           clu.NumSmartContracts() * 2,
//		"active_committee":    clu.NumSmartContracts(),
//		"dismissed_committee": 0,
//		"request_in":          0,
//		"request_out":         0,
//		"state":               0,
//	})
//	check(err, t)
//
//	for _, sc := range clu.SmartContractConfig {
//		_, err = PutChainRecord(clu, &sc)
//		check(err, t)
//	}
//
//	err = ActivateAllSC(clu)
//	check(err, t)
//
//	if !clu.WaitUntilExpectationsMet() {
//		t.Fail()
//	}
//}
//
//func TestDeactivate1Chain(t *testing.T) {
//	clu := setup(t, "test_cluster", "TestDeactivate1SC")
//
//	err := clu.ListenToMessages(map[string]int{
//		"chainrec":           3,
//		"active_committee":    1,
//		"dismissed_committee": 1,
//		"request_in":          0,
//		"request_out":         0,
//		"state":               0,
//	})
//	check(err, t)
//
//	sc := &clu.SmartContractConfig[0]
//
//	_, err = PutChainRecord(clu, sc)
//	check(err, t)
//
//	err = Activate1Chain(clu, sc)
//	check(err, t)
//
//	time.Sleep(5 * time.Second)
//
//	err = Deactivate1Chain(clu, sc)
//	check(err, t)
//
//	if !clu.WaitUntilExpectationsMet() {
//		t.Fail()
//	}
//}
