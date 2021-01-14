package testcore

import (
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core/testcore/test_sandbox"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGetSet(t *testing.T) {
	_, chain, _ := setupForTestSandbox(t)

	req := solo.NewCall(test_sandbox.Interface.Name, test_sandbox.FuncSetInt,
		test_sandbox.ParamIntParamName, "ppp",
		test_sandbox.ParamIntParamValue, 314,
	)
	_, err := chain.PostRequest(req, nil)
	require.NoError(t, err)

	ret, err := chain.CallView(test_sandbox.Interface.Name, test_sandbox.FuncGetInt,
		test_sandbox.ParamIntParamName, "ppp")
	require.NoError(t, err)

	retInt, exists, err := codec.DecodeInt64(ret.MustGet("ppp"))
	require.NoError(t, err)
	require.True(t, exists)
	require.EqualValues(t, 314, retInt)
}

func TestCallRecursive(t *testing.T) {
	_, chain, cID := setupForTestSandbox(t)

	req := solo.NewCall(test_sandbox.Interface.Name, test_sandbox.FuncCallOnChain,
		test_sandbox.ParamCallOption, "co",
		test_sandbox.ParamCallDepth, 5,
		test_sandbox.ParamHname, cID.Hname(),
	)
	_, err := chain.PostRequest(req, nil)
	require.NoError(t, err)
}
