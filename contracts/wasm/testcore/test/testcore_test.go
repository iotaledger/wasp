package test

import (
	"fmt"
	"testing"

	"github.com/iotaledger/wasp/contracts/wasm/testcore/go/testcore"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core/testcore/sbtests/sbtestsc"
	"github.com/iotaledger/wasp/wasmvm/wasmlib/go/wasmlib"
	"github.com/iotaledger/wasp/wasmvm/wasmlib/go/wasmlib/coreaccounts"
	"github.com/iotaledger/wasp/wasmvm/wasmlib/go/wasmlib/coregovernance"
	"github.com/iotaledger/wasp/wasmvm/wasmlib/go/wasmlib/coreroot"
	wasmsolo2 "github.com/iotaledger/wasp/packages/wasmvm/wasmsolo"
	"github.com/stretchr/testify/require"
)

func deployTestCore(t *testing.T, runWasm bool, addCreator ...bool) *wasmsolo2.SoloContext {
	chain := wasmsolo2.StartChain(t, "chain1")

	var creator *wasmsolo2.SoloAgent
	if len(addCreator) != 0 && addCreator[0] {
		creator = wasmsolo2.NewSoloAgent(chain.Env)
		setDeployer(t, &wasmsolo2.SoloContext{Chain: chain}, creator)
	}

	ctx := deployTestCoreOnChain(t, runWasm, chain, creator)
	require.NoError(t, ctx.Err)
	return ctx
}

func deployTestCoreOnChain(t *testing.T, runWasm bool, chain *solo.Chain, creator *wasmsolo2.SoloAgent, init ...*wasmlib.ScInitFunc) *wasmsolo2.SoloContext {
	if runWasm {
		return wasmsolo2.NewSoloContextForChain(t, chain, creator, testcore.ScName, testcore.OnLoad, init...)
	}

	return wasmsolo2.NewSoloContextForNative(t, chain, creator, testcore.ScName, testcore.OnLoad, sbtestsc.Processor, init...)
}

func run2(t *testing.T, test func(*testing.T, bool), skipWasm ...bool) {
	t.Run(fmt.Sprintf("run CORE version of %s", t.Name()), func(t *testing.T) {
		test(t, false)
	})

	if len(skipWasm) != 0 && skipWasm[0] {
		t.Logf("skipped Wasm versions of '%s'", t.Name())
		return
	}

	saveGoDebug := *wasmsolo2.GoDebug
	saveGoWasm := *wasmsolo2.GoWasm
	saveTsWasm := *wasmsolo2.TsWasm
	*wasmsolo2.GoDebug = false
	*wasmsolo2.GoWasm = false
	*wasmsolo2.TsWasm = false

	//exists, _ := util.ExistsFilePath("../pkg/testcore_bg.wasm")
	//if exists {
	//	wasmlib.ConnectHost(nil)
	//	t.Run(fmt.Sprintf("run RUST version of %s", t.Name()), func(t *testing.T) {
	//		test(t, true)
	//	})
	//}

	exists, _ := util.ExistsFilePath("../go/pkg/testcore_go.wasm")
	if exists {
		*wasmsolo2.GoWasm = true
		wasmlib.ConnectHost(nil)
		t.Run(fmt.Sprintf("run GO version of %s", t.Name()), func(t *testing.T) {
			test(t, true)
		})
		*wasmsolo2.GoWasm = false
	}

	//exists, _ = util.ExistsFilePath("../ts/pkg/testcore_ts.wasm")
	//if exists {
	//	*wasmsolo.TsWasm = true
	//	wasmlib.ConnectHost(nil)
	//	t.Run(fmt.Sprintf("run TS version of %s", t.Name()), func(t *testing.T) {
	//		test(t, true)
	//	})
	//	*wasmsolo.TsWasm = false
	//}

	*wasmsolo2.GoDebug = true
	wasmlib.ConnectHost(nil)
	t.Run(fmt.Sprintf("run GOVM version of %s", t.Name()), func(t *testing.T) {
		test(t, true)
	})

	*wasmsolo2.GoDebug = saveGoDebug
	*wasmsolo2.GoWasm = saveGoWasm
	*wasmsolo2.TsWasm = saveTsWasm
}

func TestDeployTestCore(t *testing.T) {
	run2(t, func(t *testing.T, w bool) {
		ctx := deployTestCore(t, w)
		require.EqualValues(t, ctx.Originator(), ctx.Creator())
	})
}

func TestDeployTestCoreWithCreator(t *testing.T) {
	run2(t, func(t *testing.T, w bool) {
		ctx := deployTestCore(t, w, true)
		require.NotEqualValues(t, ctx.Originator(), ctx.Creator())
	})
}

// chainAccountBalances checks the balance of the chain account and the total
// balance of all accounts, taking any extra uploadWasm() into account
func chainAccountBalances(ctx *wasmsolo2.SoloContext, w bool, chain, total uint64) {
	if w {
		// wasm setup takes 1 more iota than core setup due to uploadWasm()
		chain++
		total++
	}
	ctx.Chain.AssertCommonAccountIotas(chain)
	ctx.Chain.AssertTotalIotas(total)
}

// originatorBalanceReducedBy checks the balance of the originator address has
// reduced by the given amount, taking any extra uploadWasm() into account
func originatorBalanceReducedBy(ctx *wasmsolo2.SoloContext, w bool, minus uint64) {
	if w {
		// wasm setup takes 1 more iota than core setup due to uploadWasm()
		minus++
	}
	ctx.Chain.Env.AssertAddressIotas(ctx.Chain.OriginatorAddress, solo.Saldo-solo.ChainDustThreshold-minus)
}

func deposit(t *testing.T, ctx *wasmsolo2.SoloContext, user, target *wasmsolo2.SoloAgent, amount uint64) {
	ctxAcc := ctx.SoloContextForCore(t, coreaccounts.ScName, coreaccounts.OnLoad)
	f := coreaccounts.ScFuncs.Deposit(ctxAcc.Sign(user))
	if target != nil {
		f.Params.AgentID().SetValue(target.ScAgentID())
	}
	f.Func.TransferIotas(amount).Post()
	require.NoError(t, ctxAcc.Err)
}

func setDeployer(t *testing.T, ctx *wasmsolo2.SoloContext, deployer *wasmsolo2.SoloAgent) {
	ctxRoot := ctx.SoloContextForCore(t, coreroot.ScName, coreroot.OnLoad)
	f := coreroot.ScFuncs.GrantDeployPermission(ctxRoot)
	f.Params.Deployer().SetValue(deployer.ScAgentID())
	f.Func.TransferIotas(1).Post()
	require.NoError(t, ctxRoot.Err)
}

func setOwnerFee(t *testing.T, ctx *wasmsolo2.SoloContext, amount int64) {
	ctxGov := ctx.SoloContextForCore(t, coregovernance.ScName, coregovernance.OnLoad)
	f := coregovernance.ScFuncs.SetContractFee(ctxGov)
	f.Params.Hname().SetValue(testcore.HScName)
	f.Params.OwnerFee().SetValue(amount)
	f.Func.TransferIotas(1).Post()
	require.NoError(t, ctxGov.Err)
}

func withdraw(t *testing.T, ctx *wasmsolo2.SoloContext, user *wasmsolo2.SoloAgent) {
	ctxAcc := ctx.SoloContextForCore(t, coreaccounts.ScName, coreaccounts.OnLoad)
	f := coreaccounts.ScFuncs.Withdraw(ctxAcc.Sign(user))
	f.Func.TransferIotas(1).Post()
	require.NoError(t, ctxAcc.Err)
}
