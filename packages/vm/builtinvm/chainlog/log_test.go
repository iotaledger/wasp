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
	glb := alone.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")

	chain.CheckBase()

	err := chain.DeployContract(nil, Interface.Name, Interface.ProgramHash)

	require.NoError(t, err)
}

func TestStore(t *testing.T) {
	glb := alone.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	err := chain.DeployContract(nil, Interface.Name, Interface.ProgramHash)
	require.NoError(t, err)

	req := alone.NewCall(Interface.Name, FuncStoreLog, ParamLog, []byte("some test text"))

	_, err = chain.PostRequest(req, nil)

	require.NoError(t, err)
	res, err := chain.CallView(Interface.Name, FuncGetLog)
	require.NoError(t, err)

	v, ok, err := codec.DecodeInt64(res.MustGet(VarLogName))

	require.NoError(t, err)
	require.True(t, ok)
	require.EqualValues(t, 1, v)
}

func TestGetLasts3(t *testing.T) {
	glb := alone.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	err := chain.DeployContract(nil, Interface.Name, Interface.ProgramHash)
	require.NoError(t, err)

	req := alone.NewCall(Interface.Name, FuncStoreLog, ParamLog, []byte("PostRequest Number ONE"))
	_, err = chain.PostRequest(req, nil)

	req2 := alone.NewCall(Interface.Name, FuncStoreLog, ParamLog, []byte("PostRequest Number TWO"))
	_, err = chain.PostRequest(req2, nil)

	req3 := alone.NewCall(Interface.Name, FuncStoreLog, ParamLog, []byte("PostRequest Number THREE"))
	_, err = chain.PostRequest(req3, nil)

	res, err := chain.CallView(Interface.Name, FuncGetLasts, ParamLog, 3)
	require.NoError(t, err)

	array := datatypes.NewMustArray(res, VarLogName)

	//For some reason i'm getting always 1 more, i will continue tomorrow
	require.EqualValues(t, 3, array.Len())

}

func TestGetBetweenTs(t *testing.T) {
	//t.SkipNow()

	glb := alone.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	glb.SetTimeStep(500 * time.Millisecond)

	err := chain.DeployContract(nil, Interface.Name, Interface.ProgramHash)
	require.NoError(t, err)

	req1 := alone.NewCall(Interface.Name, FuncStoreLog, ParamLog, []byte("PostRequest Number ONE"))
	_, err = chain.PostRequest(req1, nil)

	req2 := alone.NewCall(Interface.Name, FuncStoreLog, ParamLog, []byte("PostRequest Number TWO"))
	_, err = chain.PostRequest(req2, nil)

	req3 := alone.NewCall(Interface.Name, FuncStoreLog, ParamLog, []byte("PostRequest Number THREE"))
	_, err = chain.PostRequest(req3, nil)

	req4 := alone.NewCall(Interface.Name, FuncStoreLog, ParamLog, []byte("PostRequest Number FOUR"))
	_, err = chain.PostRequest(req4, nil)

	req5 := alone.NewCall(Interface.Name, FuncStoreLog, ParamLog, []byte("PostRequest Number FIVE"))
	_, err = chain.PostRequest(req5, nil)

	res, err := chain.CallView(Interface.Name, FuncGetLogsBetweenTs,
		ParamFromTs, 0,
		ParamToTs, chain.State.Timestamp()-int64(1000*time.Millisecond),
		ParamLastsRecords, 2)
	require.NoError(t, err)

	array := datatypes.NewMustArray(res, VarLogName)

	//Expected to have the second and third record
	// var i uint16 = 0
	// for i = 0; i < array.Len(); i++ {
	// 	data, _ := array.GetAt(i)
	// 	fmt.Println(string(data))
	// }

	require.EqualValues(t, 2, array.Len())
}
