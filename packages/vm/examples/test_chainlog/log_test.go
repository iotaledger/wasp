package test_chainlog

import (
	"bytes"
	"testing"
	"time"

	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/datatypes"
	log "github.com/iotaledger/wasp/packages/vm/builtinvm/chainlog"
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

	res, err := chain.CallView(log.Interface.Name, log.FuncLenByHnameAndTR,
		log.ParamContractHname, Interface.Hname(),
		log.ParamRecordType, log.TRGenericData,
	)
	require.NoError(t, err)

	var buf bytes.Buffer
	buf.Write(Interface.Hname().Bytes())
	buf.WriteByte(log.TRGenericData)

	v, ok, err := codec.DecodeInt64(res.MustGet(kv.Key(buf.String())))

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

	_, err = chain.CallView(log.Interface.Name, log.FuncGetLogsBetweenTs,
		log.ParamFromTs, 0,
		log.ParamToTs, chain.State.Timestamp(),
		log.ParamLastsRecords, 3,
		log.ParamContractHname, Interface.Hname(),
		log.ParamRecordType, 8,
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

	res, err := chain.CallView(log.Interface.Name, log.FuncGetLogsBetweenTs,
		log.ParamFromTs, 0,
		log.ParamToTs, chain.State.Timestamp(),
		log.ParamLastsRecords, 3,
		log.ParamContractHname, Interface.Hname(),
		log.ParamRecordType, log.TRGenericData,
	)
	require.NoError(t, err)

	var buf bytes.Buffer
	buf.Write(Interface.Hname().Bytes())
	buf.WriteByte(log.TRGenericData)

	array := datatypes.NewMustArray(res, string(buf.String()))

	require.EqualValues(t, 3, array.Len())
}

func TestGetBetweenTs(t *testing.T) {
	//t.SkipNow()

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

	res, err := chain.CallView(log.Interface.Name, log.FuncGetLogsBetweenTs,
		log.ParamFromTs, 0,
		log.ParamToTs, chain.State.Timestamp()-int64(1500*time.Millisecond),
		log.ParamLastsRecords, 2,
		log.ParamContractHname, Interface.Hname(),
		log.ParamRecordType, log.TRGenericData,
	)
	require.NoError(t, err)

	var buf bytes.Buffer
	buf.Write(Interface.Hname().Bytes())
	buf.WriteByte(log.TRGenericData)

	array := datatypes.NewMustArray(res, string(buf.String()))

	require.EqualValues(t, 2, array.Len())
}

func TestGetBetweenTsAndDiferentsTypes(t *testing.T) {
	//t.SkipNow()

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

	res, err := chain.CallView(log.Interface.Name, log.FuncGetLogsBetweenTs,
		log.ParamFromTs, 0,
		log.ParamToTs, chain.State.Timestamp(),
		log.ParamLastsRecords, 3,
		log.ParamContractHname, Interface.Hname(),
		log.ParamRecordType, log.TRGenericData,
	)
	require.NoError(t, err)

	var buf bytes.Buffer
	buf.Write(Interface.Hname().Bytes())
	buf.WriteByte(log.TRGenericData)

	array := datatypes.NewMustArray(res, string(buf.String()))

	require.EqualValues(t, 3, array.Len())

}
