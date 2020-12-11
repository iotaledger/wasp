package inccounter

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/vm/alone"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

const incName = "incTest"

func checkCounter(e *alone.Chain, expected int64) {
	ret, err := e.CallView(incName, FuncGetCounter)
	require.NoError(e.Glb.T, err)
	c, ok, err := codec.DecodeInt64(ret.MustGet(VarCounter))
	require.NoError(e.Glb.T, err)
	require.True(e.Glb.T, ok)
	require.EqualValues(e.Glb.T, expected, c)
}

func TestDeployInc(t *testing.T) {
	glb := alone.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	defer chain.WaitEmptyBacklog()

	err := chain.DeployContract(nil, incName, ProgramHash)
	require.NoError(t, err)
	chain.CheckBase()
	_, _, contracts := chain.GetInfo()
	require.EqualValues(t, 4, len(contracts))
	checkCounter(chain, 0)
	chain.CheckAccountLedger()
}

func TestDeployIncInitParams(t *testing.T) {
	glb := alone.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	defer chain.WaitEmptyBacklog()

	err := chain.DeployContract(nil, incName, ProgramHash, VarCounter, 17)
	require.NoError(t, err)
	checkCounter(chain, 17)
	chain.CheckAccountLedger()
}

func TestIncDefaultParam(t *testing.T) {
	glb := alone.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	defer chain.WaitEmptyBacklog()

	err := chain.DeployContract(nil, incName, ProgramHash, VarCounter, 17)
	require.NoError(t, err)
	checkCounter(chain, 17)

	_, err = chain.PostRequest(alone.NewCall(incName, FuncIncCounter), nil)
	require.NoError(t, err)
	checkCounter(chain, 18)
	chain.CheckAccountLedger()
}

func TestIncParam(t *testing.T) {
	glb := alone.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	defer chain.WaitEmptyBacklog()

	err := chain.DeployContract(nil, incName, ProgramHash, VarCounter, 17)
	require.NoError(t, err)
	checkCounter(chain, 17)

	_, err = chain.PostRequest(alone.NewCall(incName, FuncIncCounter, VarCounter, 3), nil)
	require.NoError(t, err)
	checkCounter(chain, 20)

	chain.CheckAccountLedger()
}

func TestIncWith1Post(t *testing.T) {
	glb := alone.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")

	err := chain.DeployContract(nil, incName, ProgramHash, VarCounter, 17)
	require.NoError(t, err)
	checkCounter(chain, 17)

	req := alone.NewCall(incName, FuncIncAndRepeatOnceAfter5s).
		WithTransfer(map[balance.Color]int64{balance.ColorIOTA: 1})
	_, err = chain.PostRequest(req, nil)
	require.NoError(t, err)
	// advance logical clock to unlock that timelocked request
	glb.AdvanceClockBy(6 * time.Second)

	chain.WaitEmptyBacklog()
	checkCounter(chain, 19)

	chain.CheckAccountLedger()
}
