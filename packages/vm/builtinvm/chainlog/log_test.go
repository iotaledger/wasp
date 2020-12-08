package log

import (
	"testing"

	"github.com/iotaledger/wasp/packages/vm/alone"
	"github.com/stretchr/testify/require"
)

func TestBasic(t *testing.T) {
	e := alone.New(t, false, false)

	e.CheckBase()

	err := e.DeployContract(nil, Interface.Name, Interface.ProgramHash)

	require.NoError(t, err)
}

func TestStore(t *testing.T) {
	e := alone.New(t, false, false)
	err := e.DeployContract(nil, Interface.Name, Interface.ProgramHash)
	require.NoError(t, err)

	req := alone.NewCall(Interface.Name, FuncStoreLog, ParamLog, []byte("some test text"))

	_, err = e.PostRequest(req, nil)

	require.NoError(t, err)
}
