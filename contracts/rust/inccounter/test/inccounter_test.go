// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"fmt"
	"github.com/iotaledger/wasp/contracts/common"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/wasmhost"
	"github.com/iotaledger/wasp/packages/vm/wasmproc"
	"github.com/stretchr/testify/require"
	"sort"
	"strings"
	"testing"
	"time"
)

func setupTest(t *testing.T) *solo.Chain {
	return common.StartChainAndDeployWasmContractByName(t, ScName)
}

func TestDeploy(t *testing.T) {
	chain := common.StartChainAndDeployWasmContractByName(t, ScName)
	_, err := chain.FindContract(ScName)
	require.NoError(t, err)
}

func TestStateAfterDeploy(t *testing.T) {
	chain := common.StartChainAndDeployWasmContractByName(t, ScName)

	checkStateCounter(t, chain, nil)
}

func TestIncrementOnce(t *testing.T) {
	chain := setupTest(t)

	req := solo.NewCallParams(ScName, FuncIncrement,
	).WithIotas(1)
	_, err := chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	checkStateCounter(t, chain, 1)
}

func TestIncrementTwice(t *testing.T) {
	chain := setupTest(t)

	req := solo.NewCallParams(ScName, FuncIncrement,
	).WithIotas(1)
	_, err := chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	req = solo.NewCallParams(ScName, FuncIncrement,
	).WithIotas(1)
	_, err = chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	checkStateCounter(t, chain, 2)
}

func TestIncrementRepeatThrice(t *testing.T) {
	chain := setupTest(t)

	req := solo.NewCallParams(ScName, FuncRepeatMany,
		ParamNumRepeats, 3,
	).WithIotas(1)
	_, err := chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	chain.WaitForEmptyBacklog()

	checkStateCounter(t, chain, 4)
}

func TestIncrementCallIncrement(t *testing.T) {
	chain := setupTest(t)

	req := solo.NewCallParams(ScName, FuncCallIncrement,
	).WithIotas(1)
	_, err := chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	checkStateCounter(t, chain, 2)
}

func TestIncrementCallIncrementRecurse5x(t *testing.T) {
	chain := setupTest(t)

	req := solo.NewCallParams(ScName, FuncCallIncrementRecurse5x,
	).WithIotas(1)
	_, err := chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	checkStateCounter(t, chain, 6)
}

func TestIncrementPostIncrement(t *testing.T) {
	chain := setupTest(t)

	req := solo.NewCallParams(ScName, FuncPostIncrement,
	).WithIotas(1)
	_, err := chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	chain.WaitForEmptyBacklog()

	checkStateCounter(t, chain, 2)
}

func TestIncrementLocalStateInternalCall(t *testing.T) {
	chain := setupTest(t)

	req := solo.NewCallParams(ScName, FuncLocalStateInternalCall,
	).WithIotas(1)
	_, err := chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	checkStateCounter(t, chain, 2)
}

func TestIncrementLocalStateSandboxCall(t *testing.T) {
	chain := setupTest(t)

	req := solo.NewCallParams(ScName, FuncLocalStateSandboxCall,
	).WithIotas(1)
	_, err := chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	// global var in wasm execution has no effect
	checkStateCounter(t, chain, nil)
}

func TestIncrementLocalStatePost(t *testing.T) {
	chain := setupTest(t)

	req := solo.NewCallParams(ScName, FuncLocalStatePost,
	).WithIotas(3)
	_, err := chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	chain.WaitForEmptyBacklog()

	// global var in wasm execution has no effect
	checkStateCounter(t, chain, nil)
}

func TestLeb128(t *testing.T) {
	chain := setupTest(t)

	req := solo.NewCallParams(ScName, FuncTestLeb128,
	).WithIotas(1)
	_, err := chain.PostRequestSync(req, nil)
	require.NoError(t, err)
	res, err := chain.CallView(
		ScName, wasmproc.ViewCopyAllState,
	)
	require.NoError(t, err)
	keys := make([]string, 0)
	for key := range res {
		keys = append(keys, string(key))
	}
	sort.Strings(keys)
	for _, key := range keys {
		fmt.Printf("%s: %v\n", key, res[kv.Key(key)])
	}
}

func TestLoop(t *testing.T) {
	wasmhost.WasmTimeout = 1 * time.Second
	chain := setupTest(t)

	req := solo.NewCallParams(ScName, FuncLoop,
	).WithIotas(1)
	_, err := chain.PostRequestSync(req, nil)
	require.Error(t, err)
	errText := err.Error()
	require.True(t, strings.Contains(errText, "interrupt"))

	req = solo.NewCallParams(ScName, FuncIncrement,
	).WithIotas(1)
	_, err = chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	checkStateCounter(t, chain, 1)
}

func checkStateCounter(t *testing.T, chain *solo.Chain, expected interface{}) {
	res, err := chain.CallView(
		ScName, ViewGetCounter,
	)
	require.NoError(t, err)
	counter, exists, err := codec.DecodeInt64(res[VarCounter])
	require.NoError(t, err)
	if expected == nil {
		require.False(t, exists)
		return
	}
	require.True(t, exists)
	require.EqualValues(t, expected, counter)
}
