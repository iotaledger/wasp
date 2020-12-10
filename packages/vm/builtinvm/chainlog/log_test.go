package log

import (
	"testing"
	"time"

	"github.com/iotaledger/wasp/packages/kv/datatypes"

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

	array, err := datatypes.NewArray(res, VarLogName)
	require.NoError(t, err)

	//For some reason i'm getting always 1 more, i will continue tomorrow
	require.EqualValues(t, 3, array.Len())

}

func TestGetBetweenTs(t *testing.T) {
	t.SkipNow()

	e := alone.New(t, false, false)
	e.SetTimeStep(500 * time.Millisecond)

	err := e.DeployContract(nil, Interface.Name, Interface.ProgramHash)
	require.NoError(t, err)

	req1 := alone.NewCall(Interface.Name, FuncStoreLog, ParamLog, []byte("PostRequest Number ONE"))
	_, err = e.PostRequest(req1, nil)
	time.Sleep(500 * time.Millisecond)

	req2 := alone.NewCall(Interface.Name, FuncStoreLog, ParamLog, []byte("PostRequest Number TWO"))
	_, err = e.PostRequest(req2, nil)
	//time.Sleep(500 * time.Millisecond)

	req3 := alone.NewCall(Interface.Name, FuncStoreLog, ParamLog, []byte("PostRequest Number THREE"))
	_, err = e.PostRequest(req3, nil)
	//time.Sleep(500 * time.Millisecond)

	req4 := alone.NewCall(Interface.Name, FuncStoreLog, ParamLog, []byte("PostRequest Number FOUR"))
	_, err = e.PostRequest(req4, nil)
	//time.Sleep(500 * time.Millisecond)

	req5 := alone.NewCall(Interface.Name, FuncStoreLog, ParamLog, []byte("PostRequest Number FIVE"))
	_, err = e.PostRequest(req5, nil)
	//time.Sleep(1000 * time.Millisecond)

	res, err := e.CallView(Interface.Name, FuncGetLogsBetweenTs,
		ParamFromTs, 0,
		ParamToTs, time.Now().UnixNano()-2*time.Second.Nanoseconds(),
		ParamLastsRecords, 2)
	require.NoError(t, err)

	array, err := datatypes.NewArray(res, VarLogName)

	//Expected to have the second and third record
	// var i uint16 = 0
	// for i = 0; i < array.Len(); i++ {
	// 	data, _ := array.GetAt(i)
	// 	fmt.Println(string(data))
	// }
	require.NoError(t, err)

	require.EqualValues(t, 2, array.Len())
}
