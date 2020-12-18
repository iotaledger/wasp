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

func TestChainlogGetLenByHnameAndTR(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	err := chain.DeployContract(nil, Interface.Name, Interface.ProgramHash)
	require.NoError(t, err)

	req := solo.NewCall(Interface.Name,
		FuncChainLogGenericData,
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

func TestChainlogWrongTypeParam(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")

	err := chain.DeployContract(nil, Interface.Name, Interface.ProgramHash)
	require.NoError(t, err)

	req := solo.NewCall(Interface.Name,
		FuncChainLogGenericData,
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

func TestChainlogGetLasts3(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	glb.SetTimeStep(500 * time.Millisecond)

	err := chain.DeployContract(nil, Interface.Name, Interface.ProgramHash)
	require.NoError(t, err)

	for i := 1; i < 6; i++ {
		req := solo.NewCall(Interface.Name,
			FuncChainLogGenericData,
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

func TestChainlogGetBetweenTs(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	glb.SetTimeStep(500 * time.Millisecond)

	err := chain.DeployContract(nil, Interface.Name, Interface.ProgramHash)
	require.NoError(t, err)

	for i := 1; i < 6; i++ {
		req := solo.NewCall(Interface.Name,
			FuncChainLogGenericData,
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

func TestChainlogEventData(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")

	err := chain.DeployContract(nil, Interface.Name, Interface.ProgramHash)
	require.NoError(t, err)

	req := solo.NewCall(Interface.Name,
		FuncChainLogEventData, //Sandbox call se the TREvent type
	)
	_, err = chain.PostRequest(req, nil)

	res, err := chain.CallView(chainlog.Interface.Name, chainlog.FuncGetLogRecords,
		chainlog.ParamFromTs, 0,
		chainlog.ParamToTs, chain.State.Timestamp(),
		chainlog.ParamContractHname, Interface.Hname(),
		chainlog.ParamRecordType, chainlog.TREvent,
	)
	require.NoError(t, err)
	array := datatypes.NewMustArray(res, chainlog.ParamRecords)

	require.EqualValues(t, 1, array.Len())
}

func TestChainlogEventDataFomatted(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")

	err := chain.DeployContract(nil, Interface.Name, Interface.ProgramHash)
	require.NoError(t, err)

	req := solo.NewCall(Interface.Name,
		FuncChainLogEventDataFormatted, //Sandbox call se the TREvent type
	)
	_, err = chain.PostRequest(req, nil)

	res, err := chain.CallView(chainlog.Interface.Name, chainlog.FuncGetLogRecords,
		chainlog.ParamFromTs, 0,
		chainlog.ParamToTs, chain.State.Timestamp(),
		chainlog.ParamContractHname, Interface.Hname(),
		chainlog.ParamRecordType, chainlog.TREvent,
	)
	require.NoError(t, err)
	array := datatypes.NewMustArray(res, chainlog.ParamRecords)

	require.EqualValues(t, 1, array.Len())
}

func TestChainlogGetBetweenTsAndDifferentTypes(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	glb.SetTimeStep(500 * time.Millisecond)

	err := chain.DeployContract(nil, Interface.Name, Interface.ProgramHash)
	require.NoError(t, err)

	for i := 1; i < 4; i++ {
		req := solo.NewCall(Interface.Name,
			FuncChainLogEventData, //Sandbox call se the TREvent type
			VarCounter, i,
		)
		_, err = chain.PostRequest(req, nil)
		require.NoError(t, err)
	}
	for i := 4; i < 6; i++ {
		req := solo.NewCall(Interface.Name,
			FuncChainLogGenericData, //Sandbox call use the TRGeneric type
			VarCounter, i,
		)
		_, err = chain.PostRequest(req, nil)
		require.NoError(t, err)
	}

	//This call should return all the record that has the TREvent type (in this case 3)
	res, err := chain.CallView(chainlog.Interface.Name, chainlog.FuncGetLogRecords,
		chainlog.ParamFromTs, 0,
		chainlog.ParamToTs, chain.State.Timestamp(),
		chainlog.ParamContractHname, Interface.Hname(),
		chainlog.ParamRecordType, chainlog.TREvent,
	)
	require.NoError(t, err)
	array := datatypes.NewMustArray(res, chainlog.ParamRecords)

	require.EqualValues(t, 3, array.Len())
}

func TestChainOwnerID(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	originator := chain.OriginatorAgentID.Bytes()

	err := chain.DeployContract(nil, Interface.Name, Interface.ProgramHash)
	require.NoError(t, err)

	req := solo.NewCall(Interface.Name,
		FuncChainOwnerID,
	)
	ret, err := chain.PostRequest(req, nil)
	require.NoError(t, err)

	c := ret.MustGet(VarChainOwner)

	require.EqualValues(t, originator, c)
}

func TestChainID(t *testing.T) {
	glb := solo.New(t, false, false)

	chain := glb.NewChain(nil, "chain1")

	chainID := chain.ChainID.Bytes()

	err := chain.DeployContract(nil, Interface.Name, Interface.ProgramHash)
	require.NoError(t, err)

	req := solo.NewCall(Interface.Name,
		FuncChainID,
	)
	ret, err := chain.PostRequest(req, nil)
	require.NoError(t, err)

	c := ret.MustGet(VarChainID)

	require.EqualValues(t, chainID, c)
}

func TestSandboxCall(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	//Description of root (solo tool)
	desc := "'solo' testing chain"

	err := chain.DeployContract(nil, Interface.Name, Interface.ProgramHash)
	require.NoError(t, err)

	req := solo.NewCall(Interface.Name,
		FuncSandboxCall,
	)
	ret, err := chain.PostRequest(req, nil)
	require.NoError(t, err)

	d := ret.MustGet(VarSandboxCall)

	require.EqualValues(t, desc, string(d))
}
