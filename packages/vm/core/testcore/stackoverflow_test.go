package testcore

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/testutil/testmisc"
	"github.com/iotaledger/wasp/packages/vm"
)

func TestSandboxStackOverflow(t *testing.T) {
	contract := coreutil.NewContract("test stack overflow", "")
	testFunc := coreutil.Func("overflow")
	env := solo.New(t, &solo.InitOptions{
		AutoAdjustStorageDeposit: true,
	}).WithNativeContract(
		contract.Processor(
			func(ctx isc.Sandbox) dict.Dict { return nil },
			testFunc.WithHandler(func(ctx isc.Sandbox) dict.Dict {
				ctx.Call(contract.Hname(), testFunc.Hname(), nil, nil)
				return nil
			}),
		),
	)

	chain := env.NewChain()

	err := chain.DeployContract(nil, contract.Name, contract.ProgramHash)
	require.NoError(t, err)

	_, err = chain.PostRequestSync(solo.NewCallParams(contract.Name, testFunc.Name).WithGasBudget(math.MaxUint64), nil)
	require.Error(t, err)
	testmisc.RequireErrorToBe(t, err, vm.ErrGasBudgetExceeded)
}
