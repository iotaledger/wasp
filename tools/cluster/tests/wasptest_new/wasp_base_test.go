package wasptest

import (
	"bytes"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
	"github.com/iotaledger/wasp/packages/vm/examples"
	"github.com/iotaledger/wasp/packages/vm/examples/inccounter"
	"github.com/iotaledger/wasp/plugins/wasmtimevm"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
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

	chain.WithSCState(root.Hname, func(host string, blockIndex uint32, state codec.ImmutableMustCodec) bool {
		t.Logf("Verifying state of SC %s, node %s", root.Hname.String(), host)

		require.EqualValues(t, 1, blockIndex)

		require.EqualValues(t, []byte{0xFF}, state.Get(root.VarStateInitialized))

		chid, _ := state.GetChainID(root.VarChainID)
		require.EqualValues(t, &chain.ChainID, chid)

		desc, _ := state.GetString(root.VarDescription)
		require.EqualValues(t, chain.Description, desc)

		contractRegistry := state.GetMap(root.VarContractRegistry)
		require.EqualValues(t, 1, contractRegistry.Len())

		crBytes := contractRegistry.GetAt(root.Hname.Bytes())
		require.NotNil(t, crBytes)

		require.True(t, bytes.Equal(crBytes, util.MustBytes(root.GetRootContractRecord())))
		return true
	})
}

func TestDeployContractOnly(t *testing.T) {
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

	name := "inncounter1"
	hname := coretypes.Hn(name)
	description := "testing contract deployment with inccounter"

	_, err = chain.DeployBuiltinContract(name, examples.VMType, inccounter.ProgramHash, description, map[string]interface{}{
		inccounter.VarCounter: 42,
		root.ParamName:        name,
	})
	check(err, t)

	if !clu.WaitUntilExpectationsMet() {
		t.Fail()
	}

	chain.WithSCState(root.Hname, func(host string, blockIndex uint32, state codec.ImmutableMustCodec) bool {
		t.Logf("Verifying state of SC 0, node %s blockIndex %d", host, blockIndex)

		require.EqualValues(t, 2, blockIndex)

		contractRegistry := state.GetMap(root.VarContractRegistry)
		require.EqualValues(t, 2, contractRegistry.Len())

		crBytes := contractRegistry.GetAt(hname.Bytes())
		require.NotNil(t, crBytes)

		cr, err := root.DecodeContractRecord(crBytes)
		check(err, t)

		require.EqualValues(t, examples.VMType, cr.VMType)
		require.EqualValues(t, description, cr.Description)
		require.EqualValues(t, 0, cr.NodeFee)
		require.EqualValues(t, cr.Name, name)

		return true
	})
	chain.WithSCState(hname, func(host string, blockIndex uint32, state codec.ImmutableMustCodec) bool {
		t.Logf("Verifying state of SC %s, node %s blockIndex %d", hname.String(), host, blockIndex)

		counterValue, _ := state.GetInt64(inccounter.VarCounter)
		require.EqualValues(t, 42, counterValue)

		return true
	})
}

func TestDeployContractAndSpawn(t *testing.T) {
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
	name := "inncounter1"
	hname := coretypes.Hn(name)

	_, err = chain.DeployBuiltinContract(name, examples.VMType, inccounter.ProgramHash, description, map[string]interface{}{
		inccounter.VarCounter: 42,
	})
	check(err, t)

	if !clu.WaitUntilExpectationsMet() {
		t.Fail()
	}

	chain.WithSCState(root.Hname, func(host string, blockIndex uint32, state codec.ImmutableMustCodec) bool {
		t.Logf("Verifying state of SC %s, node %s blockIndex %d", hname.String(), host, blockIndex)

		require.EqualValues(t, 2, blockIndex)

		contractRegistry := state.GetMap(root.VarContractRegistry)
		require.EqualValues(t, 2, contractRegistry.Len())

		crBytes := contractRegistry.GetAt(hname.Bytes())
		require.NotNil(t, crBytes)

		cr, err := root.DecodeContractRecord(crBytes)
		check(err, t)

		require.EqualValues(t, examples.VMType, cr.VMType)
		require.EqualValues(t, description, cr.Description)
		require.EqualValues(t, 0, cr.NodeFee)
		require.EqualValues(t, cr.Name, name)

		return true
	})
	chain.WithSCState(hname, func(host string, blockIndex uint32, state codec.ImmutableMustCodec) bool {
		t.Logf("Verifying state of SC 1, node %s blockIndex %d", host, blockIndex)

		counterValue, _ := state.GetInt64(inccounter.VarCounter)
		require.EqualValues(t, 42, counterValue)

		return true
	})

	// send 'spawn' request to the SC which was just deployed
	tx, err := chain.OwnerClient().PostRequest(1, inccounter.EntryPointSpawn, nil, nil, nil)
	check(err, t)

	err = chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(tx, 30*time.Second)
	check(err, t)

	chain.WithSCState(2, func(host string, blockIndex uint32, state codec.ImmutableMustCodec) bool {
		t.Logf("Verifying state of SC 2, node %s blockIndex %d", host, blockIndex)

		counterValue, _ := state.GetInt64(inccounter.VarCounter)
		require.EqualValues(t, 44, counterValue)

		return true
	})

}

func TestDeployExternalContractOnly(t *testing.T) {
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

	wasmName := "increment"
	hname := coretypes.Hn(wasmName)
	description := "Wasm PoC increment"
	err = loadWasmIntoWasps(chain, wasmName, description, map[string]interface{}{
		inccounter.VarCounter: 42,
	})
	check(err, t)

	if !clu.WaitUntilExpectationsMet() {
		t.Fail()
	}

	chain.WithSCState(root.Hname, func(host string, blockIndex uint32, state codec.ImmutableMustCodec) bool {
		t.Logf("Verifying state of SC 0, node %s blockIndex %d", host, blockIndex)

		require.EqualValues(t, 2, blockIndex)

		contractRegistry := state.GetMap(root.VarContractRegistry)
		require.EqualValues(t, 2, contractRegistry.Len())

		crBytes := contractRegistry.GetAt(hname.Bytes())
		require.NotNil(t, crBytes)

		cr, err := root.DecodeContractRecord(crBytes)
		check(err, t)

		require.EqualValues(t, cr.VMType, wasmtimevm.PluginName)
		require.EqualValues(t, cr.Name, wasmName)
		require.EqualValues(t, cr.Description, description)
		require.EqualValues(t, cr.NodeFee, 0)

		return true
	})
	chain.WithSCState(1, func(host string, blockIndex uint32, state codec.ImmutableMustCodec) bool {
		t.Logf("Verifying state of SC 1, node %s blockIndex %d", host, blockIndex)

		counterValue, _ := state.GetInt64(inccounter.VarCounter)
		require.EqualValues(t, 42, counterValue)

		return true

	})
}
