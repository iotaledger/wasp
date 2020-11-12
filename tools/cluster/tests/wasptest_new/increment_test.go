package wasptest

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
	"github.com/iotaledger/wasp/packages/vm/examples/inccounter"
	"github.com/iotaledger/wasp/plugins/wasmtimevm"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

const incName = "increment"
const incDescription = "Increment, a PoC smart contract"

func TestIncDeployment(t *testing.T) {
	clu, chain := setupAndLoad(t, incName, incDescription, 0, nil)

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

		require.EqualValues(t, wasmtimevm.PluginName, cr.VMType)
		require.EqualValues(t, incName, cr.Name)
		require.EqualValues(t, incDescription, cr.Description)
		require.EqualValues(t, 0, cr.NodeFee)

		return true
	})
	chain.WithSCState(1, func(host string, blockIndex uint32, state codec.ImmutableMustCodec) bool {
		t.Logf("Verifying state of SC 1, node %s blockIndex %d", host, blockIndex)

		counterValue, _ := state.GetInt64(inccounter.VarCounter)
		require.EqualValues(t, 0, counterValue)

		return true

	})
}

func TestIncNothing(t *testing.T) {
	testNothing(t, 1)
}

func TestInc5xNothing(t *testing.T) {
	testNothing(t, 5)
}

func testNothing(t *testing.T, numRequests int) {
	clu, chain := setupAndLoad(t, incName, incDescription, numRequests, nil)

	entryPoint := coretypes.Hn("nothing")
	for i := 0; i < numRequests; i++ {
		tx, err := chain.OwnerClient().PostRequest(1, entryPoint, nil, nil, nil)
		check(err, t)
		err = chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(tx, 30*time.Second)
		check(err, t)
	}

	if !clu.WaitUntilExpectationsMet() {
		t.Fail()
	}

	chain.WithSCState(0, func(host string, blockIndex uint32, state codec.ImmutableMustCodec) bool {
		t.Logf("Verifying state of SC 0, node %s blockIndex %d", host, blockIndex)

		require.EqualValues(t, 2+numRequests, blockIndex)

		contractRegistry := state.GetArray(root.VarContractRegistry)
		require.EqualValues(t, 2, contractRegistry.Len())

		crBytes := contractRegistry.GetAt(1)
		cr, err := root.DecodeContractRecord(crBytes)
		check(err, t)

		require.EqualValues(t, wasmtimevm.PluginName, cr.VMType)
		require.EqualValues(t, incName, cr.Name)
		require.EqualValues(t, incDescription, cr.Description)
		require.EqualValues(t, 0, cr.NodeFee)

		return true
	})
	chain.WithSCState(1, func(host string, blockIndex uint32, state codec.ImmutableMustCodec) bool {
		t.Logf("Verifying state of SC 1, node %s blockIndex %d", host, blockIndex)

		counterValue, _ := state.GetInt64(inccounter.VarCounter)
		require.EqualValues(t, 0, counterValue)

		return true
	})
}

func TestIncIncrement(t *testing.T) {
	testIncrement(t, 1)
}

func TestInc5xIncrement(t *testing.T) {
	testIncrement(t, 5)
}

func testIncrement(t *testing.T, numRequests int) {
	clu, chain := setupAndLoad(t, incName, incDescription, numRequests, nil)

	entryPoint := coretypes.Hn("increment")
	for i := 0; i < numRequests; i++ {
		tx, err := chain.OwnerClient().PostRequest(1, entryPoint, nil, nil, nil)
		check(err, t)
		err = chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(tx, 30*time.Second)
		check(err, t)
	}

	if !clu.WaitUntilExpectationsMet() {
		t.Fail()
	}

	chain.WithSCState(0, func(host string, blockIndex uint32, state codec.ImmutableMustCodec) bool {
		t.Logf("Verifying state of SC 0, node %s blockIndex %d", host, blockIndex)

		require.EqualValues(t, 2+numRequests, blockIndex)

		contractRegistry := state.GetArray(root.VarContractRegistry)
		require.EqualValues(t, 2, contractRegistry.Len())

		crBytes := contractRegistry.GetAt(1)
		cr, err := root.DecodeContractRecord(crBytes)
		check(err, t)

		require.EqualValues(t, wasmtimevm.PluginName, cr.VMType)
		require.EqualValues(t, incName, cr.Name)
		require.EqualValues(t, incDescription, cr.Description)
		require.EqualValues(t, 0, cr.NodeFee)

		return true
	})
	chain.WithSCState(1, func(host string, blockIndex uint32, state codec.ImmutableMustCodec) bool {
		t.Logf("Verifying state of SC 1, node %s blockIndex %d", host, blockIndex)

		counterValue, _ := state.GetInt64(inccounter.VarCounter)
		require.EqualValues(t, numRequests, counterValue)

		return true
	})
}

func TestIncRepeatIncrement(t *testing.T) {
	clu, chain := setupAndLoad(t, incName, incDescription, 2, nil)

	entryPoint := coretypes.Hn("incrementRepeat1")
	tx, err := chain.OwnerClient().PostRequest(1, entryPoint, nil, nil, nil)
	check(err, t)
	err = chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(tx, 30*time.Second)
	check(err, t)

	if !clu.WaitUntilExpectationsMet() {
		t.Fail()
	}

	chain.WithSCState(0, func(host string, blockIndex uint32, state codec.ImmutableMustCodec) bool {
		t.Logf("Verifying state of SC 0, node %s blockIndex %d", host, blockIndex)

		require.EqualValues(t, 3, blockIndex)

		contractRegistry := state.GetArray(root.VarContractRegistry)
		require.EqualValues(t, 2, contractRegistry.Len())

		crBytes := contractRegistry.GetAt(1)
		cr, err := root.DecodeContractRecord(crBytes)
		check(err, t)

		require.EqualValues(t, wasmtimevm.PluginName, cr.VMType)
		require.EqualValues(t, incName, cr.Name)
		require.EqualValues(t, incDescription, cr.Description)
		require.EqualValues(t, 0, cr.NodeFee)

		return true
	})
	chain.WithSCState(1, func(host string, blockIndex uint32, state codec.ImmutableMustCodec) bool {
		t.Logf("Verifying state of SC 1, node %s blockIndex %d", host, blockIndex)

		counterValue, _ := state.GetInt64(inccounter.VarCounter)
		require.EqualValues(t, 2, counterValue)

		return true
	})
}

//func TestIncRepeatManyIncrement(t *testing.T) {
//	const numRepeats = 5
//
//	wasps := setup(t, "TestIncRepeatManyIncrement")
//
//	chain, err := wasps.DeployDefaultChain()
//	check(err, t)
//	err = loadWasmIntoWasps(chain, incWasmPath, incDescription, nil)
//	check(err, t)
//
//	err = requestFunds(wasps, scOwnerAddr, "sc owner")
//	check(err, t)
//
//	err = wasps.ListenToMessages(map[string]int{
//		"chainrec":            2,
//		"active_committee":    1,
//		"dismissed_committee": 0,
//		"request_in":          1 + 1 + numRepeats,
//		"request_out":         2 + 1 + numRepeats,
//		"state":               -1,
//		"vmmsg":               -1,
//	})
//	check(err, t)
//
//	scChain, scAddr, scColor, err := startSmartContract(wasps, inccounter.ProgramHash, incDescription)
//	checkSuccess(err, t, "smart contract has been created and activated")
//
//	err = wasptest.SendSimpleRequest(wasps, scOwner.SigScheme(), waspapi.RequestBlockParams{
//		TargetContractID: coretypes.NewContractID(*scChain, 0),
//		EntryPointCode:   incCodeIncrementRepeatMany,
//		Vars: map[string]interface{}{
//			"numRepeats": numRepeats,
//		},
//		// also send 5i to the SC address to use as request tokens
//		Transfer: map[balance.Color]int64{
//			balance.ColorIOTA: 5,
//		},
//	})
//	check(err, t)
//
//	if !wasps.WaitUntilExpectationsMet() {
//		t.Fail()
//	}
//
//	if !wasps.VerifyAddressBalances(scOwnerAddr, testutil.RequestFundsAmount-6, map[balance.Color]int64{
//		balance.ColorIOTA: testutil.RequestFundsAmount - 6,
//	}, "sc owner in the end") {
//		t.Fail()
//		return
//	}
//
//	if !wasps.VerifyAddressBalances(scAddr, 6, map[balance.Color]int64{
//		balance.ColorIOTA: 5,
//		*scColor:          1,
//	}, "sc in the end") {
//		t.Fail()
//		return
//	}
//
//	if !wasps.VerifySCStateVariables2(scAddr, map[kv.Key]interface{}{
//		vmconst.VarNameOwnerAddress: scOwnerAddr[:],
//		vmconst.VarNameProgramData:  programHash[:],
//		vmconst.VarNameDescription:  incDescription,
//		"counter":                   util.Uint64To8Bytes(uint64(numRepeats + 1)),
//		"numRepeats":                util.Uint64To8Bytes(0),
//	}) {
//		t.Fail()
//	}
//}
