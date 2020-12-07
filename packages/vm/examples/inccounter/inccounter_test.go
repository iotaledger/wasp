package inccounter

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/vm/alone"
	"github.com/stretchr/testify/require"
	"testing"
)

const incName = "incTest"

func checkCounter(e *alone.AloneEnvironment, expected int64) {
	req := alone.NewCall(incName, FuncGetCounter)
	ret, err := e.PostRequest(req, nil)
	require.NoError(e.T, err)
	c, ok, err := codec.DecodeInt64(ret.MustGet(VarCounter))
	require.NoError(e.T, err)
	require.True(e.T, ok)
	require.EqualValues(e.T, expected, c)
}

func TestDeployInc(t *testing.T) {
	e := alone.New(t, false, false)
	defer e.WaitEmptyBacklog()

	err := e.DeployContract(nil, incName, ProgramHash)
	require.NoError(t, err)
	e.CheckBase()
	_, _, contracts := e.GetInfo()
	require.EqualValues(t, 4, len(contracts))
	checkCounter(e, 0)
}

func TestDeployIncInitParams(t *testing.T) {
	e := alone.New(t, false, false)
	defer e.WaitEmptyBacklog()

	err := e.DeployContract(nil, incName, ProgramHash, VarCounter, 17)
	require.NoError(t, err)
	checkCounter(e, 17)
}

func TestIncDefaultParam(t *testing.T) {
	e := alone.New(t, false, false)
	defer e.WaitEmptyBacklog()

	err := e.DeployContract(nil, incName, ProgramHash, VarCounter, 17)
	require.NoError(t, err)
	checkCounter(e, 17)

	_, err = e.PostRequest(alone.NewCall(incName, FuncIncCounter), nil)
	require.NoError(t, err)
	checkCounter(e, 18)
}

func TestIncParam(t *testing.T) {
	e := alone.New(t, false, false)
	defer e.WaitEmptyBacklog()

	err := e.DeployContract(nil, incName, ProgramHash, VarCounter, 17)
	require.NoError(t, err)
	checkCounter(e, 17)

	_, err = e.PostRequest(alone.NewCall(incName, FuncIncCounter, VarCounter, 3), nil)
	require.NoError(t, err)
	checkCounter(e, 20)
}

func TestIncWith1Post(t *testing.T) {
	e := alone.New(t, true, false)

	err := e.DeployContract(nil, incName, ProgramHash, VarCounter, 17)
	require.NoError(t, err)
	checkCounter(e, 17)

	req := alone.NewCall(incName, FuncIncAndRepeatOnceAfter5s).
		WithTransfer(map[balance.Color]int64{balance.ColorIOTA: 1})
	_, err = e.PostRequest(req, nil)
	require.NoError(t, err)

	e.WaitEmptyBacklog()
	checkCounter(e, 19)
}
