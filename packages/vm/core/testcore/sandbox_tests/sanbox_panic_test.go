package sandbox_tests

import (
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core/testcore/test_sandbox"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestPanicFull(t *testing.T) {
	_, chain := setupChain(t, nil)
	setupSC(t, chain, nil)

	req := solo.NewCall(test_sandbox.Interface.Name, test_sandbox.FuncPanicFullEP)
	_, err := chain.PostRequest(req, nil)
	require.Error(t, err)
	require.EqualValues(t, 1, strings.Count(err.Error(), test_sandbox.MsgFullPanic))
}

func TestPanicViewCall(t *testing.T) {
	_, chain := setupChain(t, nil)
	setupSC(t, chain, nil)

	_, err := chain.CallView(test_sandbox.Interface.Name, test_sandbox.FuncPanicViewEP)
	require.Error(t, err)
	require.EqualValues(t, 1, strings.Count(err.Error(), test_sandbox.MsgViewPanic))
}

func TestCallPanicFull(t *testing.T) {
	_, chain := setupChain(t, nil)
	setupSC(t, chain, nil)

	req := solo.NewCall(test_sandbox.Interface.Name, test_sandbox.FuncCallPanicFullEP)
	_, err := chain.PostRequest(req, nil)
	require.Error(t, err)
	require.EqualValues(t, 1, strings.Count(err.Error(), test_sandbox.MsgFullPanic))
}

func TestCallPanicViewFromFull(t *testing.T) {
	_, chain := setupChain(t, nil)
	setupSC(t, chain, nil)

	req := solo.NewCall(test_sandbox.Interface.Name, test_sandbox.FuncCallPanicViewEPFromFull)
	_, err := chain.PostRequest(req, nil)
	require.Error(t, err)
	require.EqualValues(t, 1, strings.Count(err.Error(), test_sandbox.MsgViewPanic))
}

func TestCallPanicViewFromView(t *testing.T) {
	_, chain := setupChain(t, nil)
	setupSC(t, chain, nil)

	_, err := chain.CallView(test_sandbox.Interface.Name, test_sandbox.FuncCallPanicViewEPFromView)
	require.Error(t, err)
	require.EqualValues(t, 1, strings.Count(err.Error(), test_sandbox.MsgViewPanic))
}
