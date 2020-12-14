package log

import (
	"testing"
	"time"

	"github.com/iotaledger/wasp/packages/kv"
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

	req := alone.NewCall(Interface.Name,
		FuncStoreLog,
		ParamLog, []byte("some test text"),
		ParamContractHname, Interface.Hname(),
		ParamType, _GENERIC_DATA,
	)

	_, err = chain.PostRequest(req, nil)
	require.NoError(t, err)

	//ParamContractHname could be any contract, in this case we use the same chainlog contract
	res, err := chain.CallView(Interface.Name, FuncGetLog,
		ParamContractHname, Interface.Hname(),
		ParamType, _GENERIC_DATA,
	)
	require.NoError(t, err)

	entry := append(Interface.Hname().Bytes(), byte(_GENERIC_DATA))

	v, ok, err := codec.DecodeInt64(res.MustGet(kv.Key(entry)))

	require.NoError(t, err)
	require.True(t, ok)
	require.EqualValues(t, 1, v)
}

func TestStoreWrongTypeParam(t *testing.T) {
	glb := alone.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	err := chain.DeployContract(nil, Interface.Name, Interface.ProgramHash)
	require.NoError(t, err)

	req := alone.NewCall(Interface.Name,
		FuncStoreLog,
		ParamLog, []byte("some test text"),
		ParamContractHname, Interface.Hname(),
		ParamType, _TOKEN_TRANSFER,
	)

	_, err = chain.PostRequest(req, nil)
	require.NoError(t, err)

	//ParamContractHname could be any contract, in this case we use the same chainlog contract
	_, err = chain.CallView(Interface.Name, FuncGetLog,
		ParamContractHname, Interface.Hname(),
		ParamType, 8,
	)
	require.Error(t, err)
}

func TestGetLasts3(t *testing.T) {
	glb := alone.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	err := chain.DeployContract(nil, Interface.Name, Interface.ProgramHash)
	require.NoError(t, err)

	req1 := alone.NewCall(Interface.Name,
		FuncStoreLog,
		ParamLog, []byte("PostRequest Number ONE"),
		ParamContractHname, Interface.Hname(),
		ParamType, _TOKEN_TRANSFER,
	)
	_, err = chain.PostRequest(req1, nil)
	require.NoError(t, err)

	req2 := alone.NewCall(Interface.Name,
		FuncStoreLog,
		ParamLog, []byte("PostRequest Number TWO"),
		ParamContractHname, Interface.Hname(),
		ParamType, _TOKEN_TRANSFER,
	)
	_, err = chain.PostRequest(req2, nil)
	require.NoError(t, err)

	req3 := alone.NewCall(Interface.Name,
		FuncStoreLog,
		ParamLog, []byte("PostRequest Number THREE"),
		ParamContractHname, Interface.Hname(),
		ParamType, _TOKEN_TRANSFER,
	)
	_, err = chain.PostRequest(req3, nil)
	require.NoError(t, err)

	res, err := chain.CallView(Interface.Name,
		FuncGetLasts,
		ParamLastsRecords, 3,
		ParamContractHname, Interface.Hname(),
		ParamType, _TOKEN_TRANSFER,
	)
	require.NoError(t, err)

	entry := append(Interface.Hname().Bytes(), byte(_TOKEN_TRANSFER))

	array := datatypes.NewMustArray(res, string(entry))

	require.EqualValues(t, 3, array.Len())
}

func TestGetBetweenTs(t *testing.T) {
	//t.SkipNow()

	glb := alone.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	glb.SetTimeStep(500 * time.Millisecond)

	err := chain.DeployContract(nil, Interface.Name, Interface.ProgramHash)
	require.NoError(t, err)

	req1 := alone.NewCall(Interface.Name, FuncStoreLog,
		ParamLog, []byte("PostRequest Number ONE"),
		ParamContractHname, Interface.Hname(),
		ParamType, _REQUEST_FUNC,
	)
	_, err = chain.PostRequest(req1, nil)

	req2 := alone.NewCall(Interface.Name, FuncStoreLog,
		ParamLog, []byte("PostRequest Number TWO"),
		ParamContractHname, Interface.Hname(),
		ParamType, _REQUEST_FUNC,
	)
	_, err = chain.PostRequest(req2, nil)

	req3 := alone.NewCall(Interface.Name, FuncStoreLog,
		ParamLog, []byte("PostRequest Number THREE"),
		ParamContractHname, Interface.Hname(),
		ParamType, _REQUEST_FUNC,
	)
	_, err = chain.PostRequest(req3, nil)

	req4 := alone.NewCall(Interface.Name, FuncStoreLog,
		ParamLog, []byte("PostRequest Number FOUR"),
		ParamContractHname, Interface.Hname(),
		ParamType, _REQUEST_FUNC,
	)
	_, err = chain.PostRequest(req4, nil)

	req5 := alone.NewCall(Interface.Name, FuncStoreLog,
		ParamLog, []byte("PostRequest Number FIVE"),
		ParamContractHname, Interface.Hname(),
		ParamType, _REQUEST_FUNC,
	)
	_, err = chain.PostRequest(req5, nil)

	res, err := chain.CallView(Interface.Name, FuncGetLogsBetweenTs,
		ParamFromTs, 0,
		ParamToTs, chain.State.Timestamp()-int64(1000*time.Millisecond),
		ParamLastsRecords, 2,
		ParamContractHname, Interface.Hname(),
		ParamType, _REQUEST_FUNC)
	require.NoError(t, err)

	entry := append(Interface.Hname().Bytes(), byte(_REQUEST_FUNC))

	array := datatypes.NewMustArray(res, string(entry))

	require.EqualValues(t, 2, array.Len())
}

func TestGetBetweenTsAndDiferentsTypes(t *testing.T) {
	//t.SkipNow()

	glb := alone.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	glb.SetTimeStep(500 * time.Millisecond)

	err := chain.DeployContract(nil, Interface.Name, Interface.ProgramHash)
	require.NoError(t, err)

	req1 := alone.NewCall(Interface.Name, FuncStoreLog,
		ParamLog, []byte("PostRequest Number ONE"),
		ParamContractHname, Interface.Hname(),
		ParamType, _REQUEST_FUNC,
	)
	_, err = chain.PostRequest(req1, nil)

	req2 := alone.NewCall(Interface.Name, FuncStoreLog,
		ParamLog, []byte("PostRequest Number TWO"),
		ParamContractHname, Interface.Hname(),
		ParamType, _GENERIC_DATA,
	)
	_, err = chain.PostRequest(req2, nil)

	req3 := alone.NewCall(Interface.Name, FuncStoreLog,
		ParamLog, []byte("PostRequest Number THREE"),
		ParamContractHname, Interface.Hname(),
		ParamType, _GENERIC_DATA,
	)
	_, err = chain.PostRequest(req3, nil)

	req4 := alone.NewCall(Interface.Name, FuncStoreLog,
		ParamLog, []byte("PostRequest Number FOUR"),
		ParamContractHname, Interface.Hname(),
		ParamType, _GENERIC_DATA,
	)
	_, err = chain.PostRequest(req4, nil)

	req5 := alone.NewCall(Interface.Name, FuncStoreLog,
		ParamLog, []byte("PostRequest Number FIVE"),
		ParamContractHname, Interface.Hname(),
		ParamType, _REQUEST_FUNC,
	)
	_, err = chain.PostRequest(req5, nil)

	res, err := chain.CallView(Interface.Name, FuncGetLogsBetweenTs,
		ParamFromTs, 0,
		ParamToTs, chain.State.Timestamp(),
		ParamLastsRecords, 3,
		ParamContractHname, Interface.Hname(),
		ParamType, _GENERIC_DATA)
	require.NoError(t, err)

	entry := append(Interface.Hname().Bytes(), byte(_GENERIC_DATA))

	array := datatypes.NewMustArray(res, string(entry))

	require.EqualValues(t, 3, array.Len())
}
