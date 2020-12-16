package test_chainlog

import (
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

func TestStore(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	err := chain.DeployContract(nil, log.Interface.Name, log.Interface.ProgramHash)
	require.NoError(t, err)

	err = chain.DeployContract(nil, Interface.Name, Interface.ProgramHash)
	require.NoError(t, err)

	req := solo.NewCall(Interface.Name,
		FuncTestStore,
	)

	_, err = chain.PostRequest(req, nil)
	require.NoError(t, err)

	res, err := chain.CallView(log.Interface.Name, log.FuncGetLog,
		log.ParamContractHname, Interface.Hname(),
		log.ParamType, log.TR_GENERIC_DATA,
	)
	require.NoError(t, err)

	entry := append(Interface.Hname().Bytes(), byte(log.TR_GENERIC_DATA))

	v, ok, err := codec.DecodeInt64(res.MustGet(kv.Key(entry)))

	require.NoError(t, err)
	require.True(t, ok)
	require.EqualValues(t, 1, v)
}

func TestStoreWrongTypeParam(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	err := chain.DeployContract(nil, log.Interface.Name, log.Interface.ProgramHash)
	require.NoError(t, err)

	err = chain.DeployContract(nil, Interface.Name, Interface.ProgramHash)
	require.NoError(t, err)

	req := solo.NewCall(Interface.Name,
		FuncTestStore,
	)
	_, err = chain.PostRequest(req, nil)
	require.NoError(t, err)

	_, err = chain.CallView(log.Interface.Name, log.FuncGetLog,
		log.ParamContractHname, Interface.Hname(),
		log.ParamType, 8,
	)
	require.Error(t, err)
}

func TestGetLasts3(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	err := chain.DeployContract(nil, log.Interface.Name, log.Interface.ProgramHash)
	require.NoError(t, err)

	err = chain.DeployContract(nil, Interface.Name, Interface.ProgramHash)
	require.NoError(t, err)

	req := solo.NewCall(Interface.Name,
		FuncTestGetLasts3,
	)
	_, err = chain.PostRequest(req, nil)
	require.NoError(t, err)

	res, err := chain.CallView(log.Interface.Name, log.FuncGetLasts,
		log.ParamLastsRecords, 3,
		log.ParamContractHname, Interface.Hname(),
		log.ParamType, log.TR_TOKEN_TRANSFER,
	)
	require.NoError(t, err)

	entry := append(Interface.Hname().Bytes(), byte(log.TR_TOKEN_TRANSFER))

	array := datatypes.NewMustArray(res, string(entry))

	require.EqualValues(t, 3, array.Len())
}

func TestGetBetweenTs(t *testing.T) {
	//t.SkipNow()

	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	glb.SetTimeStep(500 * time.Millisecond)

	err := chain.DeployContract(nil, log.Interface.Name, log.Interface.ProgramHash)
	require.NoError(t, err)

	err = chain.DeployContract(nil, Interface.Name, Interface.ProgramHash)
	require.NoError(t, err)

	for i := 1; i < 6; i++ {
		req := solo.NewCall(Interface.Name,
			FuncTestGeneric,
			VarCounter, i,
			TypeRecord, log.TR_REQUEST_FUNC,
		)
		_, err = chain.PostRequest(req, nil)
		require.NoError(t, err)
	}

	res, err := chain.CallView(log.Interface.Name, log.FuncGetLogsBetweenTs,
		log.ParamFromTs, 0,
		log.ParamToTs, chain.State.Timestamp()-int64(1500*time.Millisecond),
		log.ParamLastsRecords, 2,
		log.ParamContractHname, Interface.Hname(),
		log.ParamType, log.TR_REQUEST_FUNC,
	)
	require.NoError(t, err)

	entry := append(Interface.Hname().Bytes(), byte(log.TR_REQUEST_FUNC))

	array := datatypes.NewMustArray(res, string(entry))

	require.EqualValues(t, 2, array.Len())
}

func TestGetBetweenTsAndDiferentsTypes(t *testing.T) {
	//t.SkipNow()

	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	glb.SetTimeStep(500 * time.Millisecond)

	err := chain.DeployContract(nil, log.Interface.Name, log.Interface.ProgramHash)
	require.NoError(t, err)

	err = chain.DeployContract(nil, Interface.Name, Interface.ProgramHash)
	require.NoError(t, err)

	for i := 1; i < 4; i++ {
		req := solo.NewCall(Interface.Name,
			FuncTestGeneric,
			VarCounter, i,
			TypeRecord, log.TR_GENERIC_DATA,
		)
		_, err = chain.PostRequest(req, nil)
		require.NoError(t, err)
	}
	for i := 4; i < 6; i++ {
		req := solo.NewCall(Interface.Name,
			FuncTestGeneric,
			VarCounter, i,
			TypeRecord, log.TR_REQUEST_FUNC,
		)
		_, err = chain.PostRequest(req, nil)
		require.NoError(t, err)
	}

	res, err := chain.CallView(log.Interface.Name, log.FuncGetLogsBetweenTs,
		log.ParamFromTs, 0,
		log.ParamToTs, chain.State.Timestamp(),
		log.ParamLastsRecords, 3,
		log.ParamContractHname, Interface.Hname(),
		log.ParamType, log.TR_GENERIC_DATA,
	)
	require.NoError(t, err)

	entry := append(Interface.Hname().Bytes(), byte(log.TR_GENERIC_DATA))

	array := datatypes.NewMustArray(res, string(entry))

	require.EqualValues(t, 3, array.Len())

}
