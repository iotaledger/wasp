package inccounter

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

const incName = "incTest"

func checkCounter(e *solo.Chain, expected int64) {
	ret, err := e.CallView(incName, FuncGetCounter)
	require.NoError(e.Env.T, err)
	c, ok, err := codec.DecodeInt64(ret.MustGet(VarCounter))
	require.NoError(e.Env.T, err)
	require.True(e.Env.T, ok)
	require.EqualValues(e.Env.T, expected, c)
}

func TestDeployInc(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")
	defer chain.WaitForEmptyBacklog()

	err := chain.DeployContract(nil, incName, Interface.ProgramHash)
	require.NoError(t, err)
	chain.CheckChain()
	_, contracts := chain.GetInfo()
	require.EqualValues(t, 5, len(contracts))
	checkCounter(chain, 0)
	chain.CheckAccountLedger()
}

func TestDeployIncInitParams(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")
	defer chain.WaitForEmptyBacklog()

	err := chain.DeployContract(nil, incName, Interface.ProgramHash, VarCounter, 17)
	require.NoError(t, err)
	checkCounter(chain, 17)
	chain.CheckAccountLedger()
}

func TestIncDefaultParam(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")
	defer chain.WaitForEmptyBacklog()

	err := chain.DeployContract(nil, incName, Interface.ProgramHash, VarCounter, 17)
	require.NoError(t, err)
	checkCounter(chain, 17)

	_, err = chain.PostRequestSync(solo.NewCallParams(incName, FuncIncCounter), nil)
	require.NoError(t, err)
	checkCounter(chain, 18)
	chain.CheckAccountLedger()
}

func TestIncParam(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")
	defer chain.WaitForEmptyBacklog()

	err := chain.DeployContract(nil, incName, Interface.ProgramHash, VarCounter, 17)
	require.NoError(t, err)
	checkCounter(chain, 17)

	_, err = chain.PostRequestSync(solo.NewCallParams(incName, FuncIncCounter, VarCounter, 3), nil)
	require.NoError(t, err)
	checkCounter(chain, 20)

	chain.CheckAccountLedger()
}

func TestIncWith1Post(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	err := chain.DeployContract(nil, incName, Interface.ProgramHash, VarCounter, 17)
	require.NoError(t, err)
	checkCounter(chain, 17)

	req := solo.NewCallParams(incName, FuncIncAndRepeatOnceAfter5s).
		WithTransfer(balance.ColorIOTA, 1)
	_, err = chain.PostRequestSync(req, nil)
	require.NoError(t, err)
	// advance logical clock to unlock that timelocked request
	env.AdvanceClockBy(6 * time.Second)

	chain.WaitForEmptyBacklog()
	checkCounter(chain, 19)

	chain.CheckAccountLedger()
}
