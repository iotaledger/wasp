package test_env

import (
	"github.com/iotaledger/wasp/packages/vm/alone"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestInitDummy1(t *testing.T) {
	e := alone.New(t, false, false)
	err := e.DeployContract(nil, "dummy", ProgramHash)
	require.NoError(t, err)
}

func TestFailInitDummy2(t *testing.T) {
	e := alone.New(t, false, false)
	err := e.DeployContract(nil, "dummy", ProgramHash, "fail", 1)
	require.Error(t, err)
}

func TestFailInitDummy3(t *testing.T) {
	e := alone.New(t, true, false)
	err := e.DeployContract(nil, "dummy", ProgramHash, "fail", 1)
	require.Error(t, err)
	_, _, rec := e.GetInfo()
	require.EqualValues(t, 3, len(rec))

	// repeat must succeed
	err = e.DeployContract(nil, "dummy", ProgramHash)
	require.NoError(t, err)
	_, _, rec = e.GetInfo()
	require.EqualValues(t, 4, len(rec))
}
