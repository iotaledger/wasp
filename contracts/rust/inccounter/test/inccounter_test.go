// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"github.com/iotaledger/wasp/contracts/testenv"
	"github.com/iotaledger/wasp/packages/vm/wasmlib"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestIncrementDeploy(t *testing.T) {
	te := testenv.NewTestEnv(t, ScName)
	checkStateCounter(te, nil)
}

func TestIncrementOnce(t *testing.T) {
	te := testenv.NewTestEnv(t, ScName)
	_ = te.NewCallParams(FuncIncrement).Post(0)
	checkStateCounter(te, 1)
}

func TestIncrementTwice(t *testing.T) {
	te := testenv.NewTestEnv(t, ScName)
	_ = te.NewCallParams(FuncIncrement).Post(0)
	_ = te.NewCallParams(FuncIncrement).Post(0)
	checkStateCounter(te, 2)
}

func TestIncrementRepeatThrice(t *testing.T) {
	te := testenv.NewTestEnv(t, ScName)
	_ = te.NewCallParams(FuncRepeatMany,
		ParamNumRepeats, 3).
		Post(1) // !!! posts to self
	te.WaitForEmptyBacklog()
	checkStateCounter(te, 4)
}

func TestIncrementCallIncrement(t *testing.T) {
	te := testenv.NewTestEnv(t, ScName)
	_ = te.NewCallParams(FuncCallIncrement).Post(0)
	checkStateCounter(te, 2)
}

func TestIncrementCallIncrementRecurse5x(t *testing.T) {
	te := testenv.NewTestEnv(t, ScName)
	_ = te.NewCallParams(FuncCallIncrementRecurse5x).Post(0)
	checkStateCounter(te, 6)
}

func TestIncrementPostIncrement(t *testing.T) {
	te := testenv.NewTestEnv(t, ScName)
	_ = te.NewCallParams(FuncPostIncrement).
		Post(1) // !!! posts to self
	te.WaitForEmptyBacklog()
	checkStateCounter(te, 2)
}

func TestIncrementLocalStateInternalCall(t *testing.T) {
	te := testenv.NewTestEnv(t, ScName)
	_ = te.NewCallParams(FuncLocalStateInternalCall).Post(0)
	checkStateCounter(te, 2)
}

func TestIncrementLocalStateSandboxCall(t *testing.T) {
	te := testenv.NewTestEnv(t, ScName)
	_ = te.NewCallParams(FuncLocalStateSandboxCall).Post(0)
	if testenv.WasmRunner == testenv.WasmRunnerGoDirect {
		// global var in direct go execution has effect
		checkStateCounter(te, 2)
		return
	}
	// global var in wasm execution has no effect
	checkStateCounter(te, nil)
}

func TestIncrementLocalStatePost(t *testing.T) {
	te := testenv.NewTestEnv(t, ScName)
	_ = te.NewCallParams(FuncLocalStatePost).
		Post(1)
	te.WaitForEmptyBacklog()
	if testenv.WasmRunner == testenv.WasmRunnerGoDirect {
		// global var in direct go execution has effect
		checkStateCounter(te, 1)
		return
	}
	// global var in wasm execution has no effect
	checkStateCounter(te, nil)
}

func TestIncrementViewCounter(t *testing.T) {
	te := testenv.NewTestEnv(t, ScName)
	_ = te.NewCallParams(FuncIncrement).Post(0)
	checkStateCounter(te, 1)

	ret := te.CallView(ViewGetCounter)
	results := te.Results(ret)
	counter := results.GetInt(wasmlib.Key(VarCounter))
	require.True(te.T, counter.Exists())
	require.EqualValues(t, 1, counter.Value())
}

func TestIncResultsTest(t *testing.T) {
	te := testenv.NewTestEnv(t, ScName)
	ret := te.NewCallParams(FuncResultsTest).Post(0)
	//ret = te.CallView( inccounter.ViewResultsCheck)
	require.EqualValues(t, 8, len(ret))
}

func TestIncStateTest(t *testing.T) {
	te := testenv.NewTestEnv(t, ScName)
	ret := te.NewCallParams(FuncStateTest).Post(0)
	ret = te.CallView(ViewStateCheck)
	require.EqualValues(t, 0, len(ret))
}

func checkStateCounter(te *testenv.TestEnv, expected interface{}) {
	state := te.State()
	counter := state.GetInt(wasmlib.Key(VarCounter))
	if expected == nil {
		require.False(te.T, counter.Exists())
		return
	}
	require.True(te.T, counter.Exists())
	require.EqualValues(te.T, expected, counter.Value())
}
