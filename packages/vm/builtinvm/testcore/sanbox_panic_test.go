package testcore

import (
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/testcore/test_sandbox"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestPanicFull(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "ch1")

	err := chain.DeployContract(nil, test_sandbox.Interface.Name, test_sandbox.Interface.ProgramHash)
	require.NoError(t, err)
	chain.CheckChain()

	req := solo.NewCall(test_sandbox.Interface.Name, test_sandbox.FuncPanicFullEP)
	_, err = chain.PostRequest(req, nil)
	require.Error(t, err)
	require.EqualValues(t, 1, strings.Count(err.Error(), test_sandbox.MsgFullPanic))
}

func TestPanicViewPost(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "ch1")

	err := chain.DeployContract(nil, test_sandbox.Interface.Name, test_sandbox.Interface.ProgramHash)
	require.NoError(t, err)
	chain.CheckChain()

	req := solo.NewCall(test_sandbox.Interface.Name, test_sandbox.FuncPanicViewEP)
	_, err = chain.PostRequest(req, nil)
	require.Error(t, err)
	require.EqualValues(t, 1, strings.Count(err.Error(), test_sandbox.MsgViewPanic))
}

func TestPanicViewCall(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "ch1")

	err := chain.DeployContract(nil, test_sandbox.Interface.Name, test_sandbox.Interface.ProgramHash)
	require.NoError(t, err)
	chain.CheckChain()

	_, err = chain.CallView(test_sandbox.Interface.Name, test_sandbox.FuncPanicViewEP)
	require.Error(t, err)
	require.EqualValues(t, 1, strings.Count(err.Error(), test_sandbox.MsgViewPanic))
}

func TestCallPanicFull(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "ch1")

	err := chain.DeployContract(nil, test_sandbox.Interface.Name, test_sandbox.Interface.ProgramHash)
	require.NoError(t, err)
	chain.CheckChain()

	req := solo.NewCall(test_sandbox.Interface.Name, test_sandbox.FuncCallPanicFullEP)
	_, err = chain.PostRequest(req, nil)
	require.Error(t, err)
	require.EqualValues(t, 1, strings.Count(err.Error(), test_sandbox.MsgFullPanic))
}

func TestCallPanicViewFromFull(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "ch1")

	err := chain.DeployContract(nil, test_sandbox.Interface.Name, test_sandbox.Interface.ProgramHash)
	require.NoError(t, err)
	chain.CheckChain()

	req := solo.NewCall(test_sandbox.Interface.Name, test_sandbox.FuncCallPanicViewEPFromFull)
	_, err = chain.PostRequest(req, nil)
	require.Error(t, err)
	require.EqualValues(t, 1, strings.Count(err.Error(), test_sandbox.MsgViewPanic))
}

func TestCallPanicViewFromView(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "ch1")

	err := chain.DeployContract(nil, test_sandbox.Interface.Name, test_sandbox.Interface.ProgramHash)
	require.NoError(t, err)
	chain.CheckChain()

	_, err = chain.CallView(test_sandbox.Interface.Name, test_sandbox.FuncCallPanicViewEPFromView)
	require.Error(t, err)
	require.EqualValues(t, 1, strings.Count(err.Error(), test_sandbox.MsgViewPanic))
}
