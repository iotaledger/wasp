package log

import (
	"github.com/iotaledger/wasp/packages/kv/codec"
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
	res, err := e.CallView(alone.NewCall(Interface.Name, FuncGetLog))
	require.NoError(t, err)

	v, ok, err := codec.DecodeInt64(res.MustGet(VarLogName))
	require.NoError(t, err)
	require.True(t, ok)
	require.EqualValues(t, 1, v)
}
