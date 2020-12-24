package test_dummy_wasm

import (
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	contractName = "test_env"
	fileName     = "test_dummy_bg.wasm"
)

func TestSuccess(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	defer chain.WaitForEmptyBacklog()

	err := chain.DeployWasmContract(nil, contractName, fileName)
	require.NoError(t, err)
	_, rec := chain.GetInfo()
	require.EqualValues(t, 5, len(rec))
}

func TestFail(t *testing.T) {
	// disabled, changes to Wasm VM but test_dummy code is no longer in wasplib
	t.SkipNow()
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	err := chain.DeployWasmContract(nil, contractName, fileName,
		"failParam", 1,
	)
	require.Error(t, err)
	_, rec := chain.GetInfo()
	require.EqualValues(t, 4, len(rec))
}

func TestFailRepeat(t *testing.T) {
	// disabled, changes to Wasm VM but test_dummy code is no longer in wasplib
	t.SkipNow()
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	err := chain.DeployWasmContract(nil, contractName, fileName,
		"failParam", 1,
	)
	require.Error(t, err)

	err = chain.DeployWasmContract(nil, contractName, fileName)
	require.NoError(t, err)
	_, rec := chain.GetInfo()
	require.EqualValues(t, 5, len(rec))
}
