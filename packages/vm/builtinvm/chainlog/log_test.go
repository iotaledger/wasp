package log

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/kv/datatypes"
	"testing"

	"github.com/iotaledger/wasp/packages/kv/codec"
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
	res, err := e.CallView(Interface.Name, FuncGetLog)
	require.NoError(t, err)

	v, ok, err := codec.DecodeInt64(res.MustGet(VarLogName))

	require.NoError(t, err)
	require.True(t, ok)
	require.EqualValues(t, 1, v)
}

func TestGetLasts3(t *testing.T) {
	e := alone.New(t, false, false)
	err := e.DeployContract(nil, Interface.Name, Interface.ProgramHash)
	require.NoError(t, err)

	req := alone.NewCall(Interface.Name, FuncStoreLog, ParamLog, []byte("PostRequest Number ONE"))
	_, err = e.PostRequest(req, nil)

	req2 := alone.NewCall(Interface.Name, FuncStoreLog, ParamLog, []byte("PostRequest Number TWO"))
	_, err = e.PostRequest(req2, nil)

	req3 := alone.NewCall(Interface.Name, FuncStoreLog, ParamLog, []byte("PostRequest Number THREE"))
	_, err = e.PostRequest(req3, nil)

	res, err := e.CallView(Interface.Name, FuncGetLasts, ParamLog, 3)
	require.NoError(t, err)

	fmt.Printf("return %s\n", res.String())

	array, err := datatypes.NewArray(res, VarLogName)
	require.NoError(t, err)

	//For some reason i'm getting always 1 more, i will continue tomorrow
	require.EqualValues(t, 3, array.Len())

}
