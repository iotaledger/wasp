package wasptest

import (
	"bytes"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
	"github.com/iotaledger/wasp/packages/vm/examples/inccounter"
	"github.com/iotaledger/wasp/plugins/wasmtimevm"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

const incName = "increment"
const incDescription = "Increment, a PoC smart contract"

var hname = coretypes.Hn(incName)

func TestIncDeployment(t *testing.T) {
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

	err = loadWasmIntoWasps(chain, incName, incDescription, nil)

	if !clu.WaitUntilExpectationsMet() {
		t.Fail()
	}

	chain.WithSCState(root.Hname, func(host string, blockIndex uint32, state codec.ImmutableMustCodec) bool {
		require.EqualValues(t, 2, blockIndex)

		contractRegistry := state.GetMap(root.VarContractRegistry)
		require.EqualValues(t, 2, contractRegistry.Len())
		//--
		crBytes := contractRegistry.GetAt(root.Hname.Bytes())
		require.NotNil(t, crBytes)
		require.True(t, bytes.Equal(crBytes, util.MustBytes(root.GetRootContractRecord())))
		//--
		crBytes = contractRegistry.GetAt(hname.Bytes())
		require.NotNil(t, crBytes)
		cr, err := root.DecodeContractRecord(crBytes)
		check(err, t)
		require.EqualValues(t, wasmtimevm.PluginName, cr.VMType)
		require.EqualValues(t, incName, cr.Name)
		require.EqualValues(t, incDescription, cr.Description)
		require.EqualValues(t, 0, cr.NodeFee)
		return true
	})
	chain.WithSCState(hname, func(host string, blockIndex uint32, state codec.ImmutableMustCodec) bool {
		counterValue, _ := state.GetInt64(inccounter.VarCounter)
		require.EqualValues(t, 0, counterValue)
		return true
	})
}

func TestIncNothing(t *testing.T) {
	testNothing(t, incName, incDescription, 1)
}

func TestInc5xNothing(t *testing.T) {
	testNothing(t, incName, incDescription, 5)
}

func testNothing(t *testing.T, name string, description string, numRequests int) {
	clu := setup(t, "test_cluster")

	err := clu.ListenToMessages(map[string]int{
		"chainrec":            2,
		"active_committee":    1,
		"dismissed_committee": 0,
		"state":               3 + numRequests,
		"request_in":          2 + numRequests,
		"request_out":         3 + numRequests,
	})
	check(err, t)

	chain, err := clu.DeployDefaultChain()
	check(err, t)

	err = loadWasmIntoWasps(chain, name, description, map[string]interface{}{
		inccounter.VarCounter: 1,
	})

	chain.WithSCState(root.Hname, func(host string, blockIndex uint32, state codec.ImmutableMustCodec) bool {
		require.EqualValues(t, 2, blockIndex)

		contractRegistry := state.GetMap(root.VarContractRegistry)
		require.EqualValues(t, 2, contractRegistry.Len())
		//--
		crBytes := contractRegistry.GetAt(root.Hname.Bytes())
		require.NotNil(t, crBytes)
		require.True(t, bytes.Equal(crBytes, util.MustBytes(root.GetRootContractRecord())))
		//--
		crBytes = contractRegistry.GetAt(hname.Bytes())
		require.NotNil(t, crBytes)
		cr, err := root.DecodeContractRecord(crBytes)
		check(err, t)
		require.EqualValues(t, wasmtimevm.PluginName, cr.VMType)
		require.EqualValues(t, incName, cr.Name)
		require.EqualValues(t, incDescription, cr.Description)
		require.EqualValues(t, 0, cr.NodeFee)
		return true
	})
	chain.WithSCState(hname, func(host string, blockIndex uint32, state codec.ImmutableMustCodec) bool {
		counterValue, _ := state.GetInt64(inccounter.VarCounter)
		require.EqualValues(t, 1, counterValue)
		return true
	})

	for i := 0; i < numRequests; i++ {
		tx, err := chain.OwnerClient().PostRequest(coretypes.Hn(name), coretypes.Hn("nothing"), nil, nil, nil)
		check(err, t)
		err = chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(tx, 30*time.Second)
		check(err, t)
	}

	if !clu.WaitUntilExpectationsMet() {
		t.Fail()
	}

	chain.WithSCState(root.Hname, func(host string, blockIndex uint32, state codec.ImmutableMustCodec) bool {
		require.EqualValues(t, numRequests+2, blockIndex)

		contractRegistry := state.GetMap(root.VarContractRegistry)
		require.EqualValues(t, 2, contractRegistry.Len())
		//--
		crBytes := contractRegistry.GetAt(root.Hname.Bytes())
		require.NotNil(t, crBytes)
		require.True(t, bytes.Equal(crBytes, util.MustBytes(root.GetRootContractRecord())))
		//--
		crBytes = contractRegistry.GetAt(hname.Bytes())
		require.NotNil(t, crBytes)
		cr, err := root.DecodeContractRecord(crBytes)
		check(err, t)
		require.EqualValues(t, wasmtimevm.PluginName, cr.VMType)
		require.EqualValues(t, incName, cr.Name)
		require.EqualValues(t, incDescription, cr.Description)
		require.EqualValues(t, 0, cr.NodeFee)
		return true
	})
	chain.WithSCState(hname, func(host string, blockIndex uint32, state codec.ImmutableMustCodec) bool {
		counterValue, _ := state.GetInt64(inccounter.VarCounter)
		require.EqualValues(t, 1, counterValue) // expected 1 ??????
		return true
	})
}

//func TestIncIncrement(t *testing.T) {
//	testIncrement(t, "TestIncIncrement", 1)
//}
//
//func TestInc5xIncrement(t *testing.T) {
//	testIncrement(t, "TestInc5xIncrement", 5)
//}
//
//func testIncrement(t *testing.T, testName string, increments int) {
//	wasps := setup(t, testName)
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
//		"request_in":          1 + increments,
//		"request_out":         2 + increments,
//		"state":               -1,
//		"vmmsg":               -1,
//	})
//	check(err, t)
//
//	scChain, scAddr, scColor, err := startSmartContract(wasps, inccounter.ProgramHash, incDescription)
//	checkSuccess(err, t, "smart contract has been created and activated")
//
//	for i := 0; i < increments; i++ {
//		err = wasptest.SendSimpleRequest(wasps, scOwner.SigScheme(), waspapi.RequestBlockParams{
//			TargetContractID: coretypes.NewContractID(*scChain, 0),
//			EntryPointCode:   incCodeIncrement,
//		})
//		check(err, t)
//	}
//
//	if !wasps.WaitUntilExpectationsMet() {
//		t.Fail()
//	}
//
//	if !wasps.VerifyAddressBalances(scOwnerAddr, testutil.RequestFundsAmount-1, map[balance.Color]int64{
//		balance.ColorIOTA: testutil.RequestFundsAmount - 1,
//	}, "sc owner in the end") {
//		t.Fail()
//		return
//	}
//
//	if !wasps.VerifyAddressBalances(scAddr, 1, map[balance.Color]int64{
//		balance.ColorIOTA: 0,
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
//		"counter":                   util.Uint64To8Bytes(uint64(increments)),
//	}) {
//		t.Fail()
//	}
//}
//
//func TestIncRepeatIncrement(t *testing.T) {
//	wasps := setup(t, "TestIncRepeatIncrement")
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
//		"request_in":          1 + 2,
//		"request_out":         2 + 2,
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
//		EntryPointCode:   incCodeIncrementRepeat1,
//		// also send 1i to the SC address to use as request token
//		Transfer: map[balance.Color]int64{
//			balance.ColorIOTA: 1,
//		},
//	})
//	check(err, t)
//
//	if !wasps.WaitUntilExpectationsMet() {
//		t.Fail()
//	}
//
//	if !wasps.VerifyAddressBalances(scOwnerAddr, testutil.RequestFundsAmount-2, map[balance.Color]int64{
//		balance.ColorIOTA: testutil.RequestFundsAmount - 2,
//	}, "sc owner in the end") {
//		t.Fail()
//		return
//	}
//
//	if !wasps.VerifyAddressBalances(scAddr, 2, map[balance.Color]int64{
//		balance.ColorIOTA: 1,
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
//		"counter":                   util.Uint64To8Bytes(uint64(2)),
//	}) {
//		t.Fail()
//	}
//}
//
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
