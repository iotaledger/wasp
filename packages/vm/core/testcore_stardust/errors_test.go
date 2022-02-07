package testcore

import (
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core"
	"github.com/stretchr/testify/require"
	"testing"
)

var errorContractName = "ErrorContract"
var errorContract = coreutil.NewContract(errorContractName, "error contract")

var funcRegisterErrors = coreutil.Func("register_errors")
var funcThrowErrors = coreutil.Func("throw_errors")

var errorContractProcessor = errorContract.Processor(nil,
	funcRegisterErrors.WithHandler(func(ctx iscp.Sandbox) dict.Dict {
		ctx.RegisterError(1, "Test Error")

		return nil
	}),
	funcThrowErrors.WithHandler(func(ctx iscp.Sandbox) dict.Dict {
		buf := make([]byte, bigEventSize)
		ctx.Event(string(buf))
		return nil
	}),
)

func TestError(t *testing.T) {
	core.PrintWellKnownHnames()
	env := solo.New(t, &solo.InitOptions{AutoAdjustDustDeposit: true})
	chain := env.NewChain(nil, "chain1")

	err := chain.DeployContract(nil, errorContract.Name, errorContract.ProgramHash)
	require.NoError(t, err)

	tx, _, err := chain.PostRequestSyncTx(
		solo.NewCallParams(errorContract.Name, funcRegisterErrors.Name).AddAssetsIotas(1),
		nil,
	)

	require.Error(t, err) // error expected (too many events)
	reqs, err := chain.Env.RequestsForChain(tx, chain.ChainID)
	require.NoError(t, err)

	t.Log(reqs)
}
