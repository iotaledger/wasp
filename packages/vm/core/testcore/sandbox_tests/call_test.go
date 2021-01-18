package sandbox_tests

import (
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core/testcore/test_sandbox"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGetSet(t *testing.T) {
	_, chain := setupChain(t, nil)
	setupTestSandboxSC(t, chain, nil)

	req := solo.NewCall(SandboxSCName, test_sandbox.FuncSetInt,
		test_sandbox.ParamIntParamName, "ppp",
		test_sandbox.ParamIntParamValue, 314,
	)
	_, err := chain.PostRequest(req, nil)
	require.NoError(t, err)

	ret, err := chain.CallView(SandboxSCName, test_sandbox.FuncGetInt,
		test_sandbox.ParamIntParamName, "ppp")
	require.NoError(t, err)

	retInt, exists, err := codec.DecodeInt64(ret.MustGet("ppp"))
	require.NoError(t, err)
	require.True(t, exists)
	require.EqualValues(t, 314, retInt)
}

func TestCallRecursive(t *testing.T) {
	if RUN_WASM {
		t.SkipNow()
	}
	_, chain := setupChain(t, nil)
	cID := setupTestSandboxSC(t, chain, nil)

	req := solo.NewCall(SandboxSCName, test_sandbox.FuncCallOnChain,
		test_sandbox.ParamCallOption, "co",
		test_sandbox.ParamIntParamValue, 50,
		test_sandbox.ParamHname, cID.Hname(),
	)
	_, err := chain.PostRequest(req, nil)
	require.NoError(t, err)
	_, err = chain.PostRequest(req, nil)
	require.NoError(t, err)
	_, err = chain.PostRequest(req, nil)
	require.NoError(t, err)
}

const n = 3

func fibo(n int64) int64 {
	if n == 0 || n == 1 {
		return n
	}
	return fibo(n-1) + fibo(n-2)
}

func TestCallFibonacci(t *testing.T) {
	if RUN_WASM {
		t.SkipNow()
	}
	_, chain := setupChain(t, nil)
	setupTestSandboxSC(t, chain, nil)

	ret, err := chain.CallView(SandboxSCName, test_sandbox.FuncGetFibonacci,
		test_sandbox.ParamIntParamValue, n,
	)
	require.NoError(t, err)
	val, exists, err := codec.DecodeInt64(ret.MustGet(test_sandbox.ParamIntParamValue))
	require.NoError(t, err)
	require.True(t, exists)
	require.EqualValues(t, fibo(n), val)
}
