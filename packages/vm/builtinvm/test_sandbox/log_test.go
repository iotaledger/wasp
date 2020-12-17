package test_sandbox

import (
	"testing"
	"time"

	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/datatypes"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/chainlog"
	"github.com/iotaledger/wasp/packages/vm/solo"
	"github.com/stretchr/testify/require"
)

func TestBasic(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")

	chain.CheckBase()
	err := chain.DeployContract(nil, Interface.Name, Interface.ProgramHash)
	require.NoError(t, err)
}

func TestGetLenByHnameAndTR(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	err := chain.DeployContract(nil, Interface.Name, Interface.ProgramHash)
	require.NoError(t, err)

	req := solo.NewCall(Interface.Name,
		FuncTestGeneric,
	)
	_, err = chain.PostRequest(req, nil)
	require.NoError(t, err)

	res, err := chain.CallView(chainlog.Interface.Name, chainlog.FuncLenByHnameAndTR,
		chainlog.ParamContractHname, Interface.Hname(),
		chainlog.ParamRecordType, chainlog.TRGenericData,
	)
	require.NoError(t, err)

	v, ok, err := codec.DecodeInt64(res.MustGet(chainlog.ParamNumRecords))

	require.NoError(t, err)
	require.True(t, ok)
	require.EqualValues(t, 1, v)
}

func TestStoreWrongTypeParam(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")

	err := chain.DeployContract(nil, Interface.Name, Interface.ProgramHash)
	require.NoError(t, err)

	req := solo.NewCall(Interface.Name,
		FuncTestGeneric,
		VarCounter, 1,
	)
	_, err = chain.PostRequest(req, nil)
	require.NoError(t, err)

	_, err = chain.CallView(chainlog.Interface.Name, chainlog.FuncGetLogRecords,
		chainlog.ParamContractHname, Interface.Hname(),
		chainlog.ParamRecordType, 8,
	)
	require.Error(t, err)
}

func TestGetLasts3(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	glb.SetTimeStep(500 * time.Millisecond)

	err := chain.DeployContract(nil, Interface.Name, Interface.ProgramHash)
	require.NoError(t, err)

	for i := 1; i < 6; i++ {
		req := solo.NewCall(Interface.Name,
			FuncTestGeneric,
			VarCounter, i,
		)
		_, err = chain.PostRequest(req, nil)
		require.NoError(t, err)
	}

	res, err := chain.CallView(chainlog.Interface.Name, chainlog.FuncGetLogRecords,
		chainlog.ParamMaxLastRecords, 3,
		chainlog.ParamContractHname, Interface.Hname(),
		chainlog.ParamRecordType, chainlog.TRGenericData,
	)
	require.NoError(t, err)

	array := datatypes.NewMustArray(res, chainlog.ParamRecords)
	require.EqualValues(t, 3, array.Len())
}

func TestGetBetweenTs(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	glb.SetTimeStep(500 * time.Millisecond)

	err := chain.DeployContract(nil, Interface.Name, Interface.ProgramHash)
	require.NoError(t, err)

	for i := 1; i < 6; i++ {
		req := solo.NewCall(Interface.Name,
			FuncTestGeneric,
			VarCounter, i,
		)
		_, err = chain.PostRequest(req, nil)
		require.NoError(t, err)
	}

	res, err := chain.CallView(chainlog.Interface.Name, chainlog.FuncGetLogRecords,
		chainlog.ParamFromTs, 0,
		chainlog.ParamToTs, chain.State.Timestamp()-int64(1500*time.Millisecond),
		chainlog.ParamMaxLastRecords, 2,
		chainlog.ParamContractHname, Interface.Hname(),
		chainlog.ParamRecordType, chainlog.TRGenericData,
	)
	require.NoError(t, err)

	array := datatypes.NewMustArray(res, chainlog.ParamRecords)
	require.EqualValues(t, 2, array.Len())
}

// not sure what is the purpose of the test
func TestGetBetweenTsAndDifferentTypes(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	glb.SetTimeStep(500 * time.Millisecond)

	err := chain.DeployContract(nil, Interface.Name, Interface.ProgramHash)
	require.NoError(t, err)

	for i := 1; i < 4; i++ {
		req := solo.NewCall(Interface.Name,
			FuncTestGeneric,
			VarCounter, i,
		)
		_, err = chain.PostRequest(req, nil)
		require.NoError(t, err)
	}
	for i := 4; i < 6; i++ {
		req := solo.NewCall(Interface.Name,
			FuncTestGeneric,
			VarCounter, i,
		)
		_, err = chain.PostRequest(req, nil)
		require.NoError(t, err)
	}

	res, err := chain.CallView(chainlog.Interface.Name, chainlog.FuncGetLogRecords,
		chainlog.ParamFromTs, 0,
		chainlog.ParamToTs, chain.State.Timestamp(),
		chainlog.ParamContractHname, Interface.Hname(),
		chainlog.ParamRecordType, chainlog.TRGenericData,
	)
	require.NoError(t, err)
	array := datatypes.NewMustArray(res, chainlog.ParamRecords)

	require.EqualValues(t, 5, array.Len())
}
