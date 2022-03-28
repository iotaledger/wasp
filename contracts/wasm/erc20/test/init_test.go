package test

import (
	"testing"

	"github.com/iotaledger/wasp/contracts/wasm/erc20/go/erc20"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core/corecontracts"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmsolo"
	"github.com/stretchr/testify/require"
)

func TestDeployErc20(t *testing.T) {
	setupTest(t)

	init := erc20.ScFuncs.Init(nil)
	init.Params.Supply().SetValue(utxodb.FundsFromFaucetAmount)
	init.Params.Creator().SetValue(creator.ScAgentID())
	ctx := wasmsolo.NewSoloContextForChain(t, chain, nil, erc20.ScName, erc20.OnLoad, init.Func)
	require.NoError(t, ctx.Err)
	_, _, rec := chain.GetInfo()
	require.EqualValues(t, len(corecontracts.All)+1, len(rec))

	require.NoError(t, ctx.ContractExists(erc20.ScName))

	// deploy second time
	init = erc20.ScFuncs.Init(nil)
	init.Params.Supply().SetValue(utxodb.FundsFromFaucetAmount)
	init.Params.Creator().SetValue(creator.ScAgentID())
	ctx = wasmsolo.NewSoloContextForChain(t, chain, nil, erc20.ScName, erc20.OnLoad, init.Func)
	require.Error(t, ctx.Err)
	_, _, rec = chain.GetInfo()
	require.EqualValues(t, len(corecontracts.All)+1, len(rec))
}

func TestDeployErc20Fail1(t *testing.T) {
	setupTest(t)
	ctx := wasmsolo.NewSoloContextForChain(t, chain, nil, erc20.ScName, erc20.OnLoad)
	require.Error(t, ctx.Err)
	_, _, rec := chain.GetInfo()
	require.EqualValues(t, len(corecontracts.All), len(rec))
}

func TestDeployErc20Fail2(t *testing.T) {
	setupTest(t)
	init := erc20.ScFuncs.Init(nil)
	init.Params.Supply().SetValue(utxodb.FundsFromFaucetAmount)
	ctx := wasmsolo.NewSoloContextForChain(t, chain, nil, erc20.ScName, erc20.OnLoad, init.Func)
	require.Error(t, ctx.Err)
	_, _, rec := chain.GetInfo()
	require.EqualValues(t, len(corecontracts.All), len(rec))
}

func TestDeployErc20Fail3(t *testing.T) {
	setupTest(t)
	init := erc20.ScFuncs.Init(nil)
	init.Params.Creator().SetValue(creator.ScAgentID())
	ctx := wasmsolo.NewSoloContextForChain(t, chain, nil, erc20.ScName, erc20.OnLoad, init.Func)
	require.Error(t, ctx.Err)
	_, _, rec := chain.GetInfo()
	require.EqualValues(t, len(corecontracts.All), len(rec))
}

func TestDeployErc20Fail3Repeat(t *testing.T) {
	setupTest(t)
	init := erc20.ScFuncs.Init(nil)
	init.Params.Creator().SetValue(creator.ScAgentID())
	ctx := wasmsolo.NewSoloContextForChain(t, chain, nil, erc20.ScName, erc20.OnLoad, init.Func)
	require.Error(t, ctx.Err)
	_, _, rec := chain.GetInfo()
	require.EqualValues(t, len(corecontracts.All), len(rec))
	require.Error(t, ctx.ContractExists(erc20.ScName))

	// repeat after failure
	init = erc20.ScFuncs.Init(nil)
	init.Params.Supply().SetValue(utxodb.FundsFromFaucetAmount)
	init.Params.Creator().SetValue(creator.ScAgentID())
	ctx = wasmsolo.NewSoloContextForChain(t, chain, nil, erc20.ScName, erc20.OnLoad, init.Func)
	require.NoError(t, ctx.Err)
	_, _, rec = chain.GetInfo()
	require.EqualValues(t, len(corecontracts.All)+1, len(rec))
	require.NoError(t, ctx.ContractExists(erc20.ScName))
}
