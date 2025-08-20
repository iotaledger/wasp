package sbtests

import (
	"testing"

	"github.com/iotaledger/wasp/v2/packages/solo"
	"github.com/iotaledger/wasp/v2/packages/testutil/testmisc"
	"github.com/iotaledger/wasp/v2/packages/vm/core/errors/coreerrors"
	"github.com/iotaledger/wasp/v2/packages/vm/core/testcore/sbtests/sbtestsc"
)

func testPanicCallView(t *testing.T, funName string, gasBudget uint64, panicView string) {
	_, chain := setupChain(t)
	setupTestSandboxSC(t, chain, nil)

	req := solo.NewCallParamsEx(ScName, funName).WithGasBudget(gasBudget)
	_, err := chain.PostRequestSync(req, nil)
	testmisc.RequireErrorToBe(t, err, panicView)
}

func TestPanicFull(t *testing.T) {
	testPanicCallView(t, sbtestsc.FuncPanicFullEP.Name, 100_000, sbtestsc.MsgFullPanic)
}

func TestPanicViewCall(t *testing.T) {
	_, chain := setupChain(t)
	setupTestSandboxSC(t, chain, nil)

	_, err := chain.CallViewEx(ScName, sbtestsc.FuncPanicViewEP.Name)
	testmisc.RequireErrorToBe(t, err, sbtestsc.MsgViewPanic)
}

func TestCallPanicFull(t *testing.T) {
	_, chain := setupChain(t)
	setupTestSandboxSC(t, chain, nil)

	req := solo.NewCallParamsEx(ScName, sbtestsc.FuncCallPanicFullEP.Name).WithGasBudget(17000)
	_, err := chain.PostRequestSync(req, nil)
	testmisc.RequireErrorToBe(t, err, coreerrors.ErrUntypedError.Create(sbtestsc.MsgFullPanic))
}

func TestCallPanicViewFromFull(t *testing.T) {
	testPanicCallView(t, sbtestsc.FuncCallPanicViewEPFromFull.Name, 20_000, sbtestsc.MsgViewPanic)
}

func TestCallPanicViewFromView(t *testing.T) {
	_, chain := setupChain(t)
	setupTestSandboxSC(t, chain, nil)

	_, err := chain.CallViewEx(ScName, sbtestsc.FuncCallPanicViewEPFromView.Name)
	testmisc.RequireErrorToBe(t, err, sbtestsc.MsgViewPanic)
}
