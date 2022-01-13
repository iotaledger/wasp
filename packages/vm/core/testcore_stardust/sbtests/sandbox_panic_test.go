package sbtests

import (
	"strings"
	"testing"

	"github.com/iotaledger/wasp/packages/testutil/testmisc"

	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core/testcore_stardust/sbtests/sbtestsc"
	"github.com/stretchr/testify/require"
)

func TestPanicFull(t *testing.T) { run2(t, testPanicFull) }
func testPanicFull(t *testing.T, w bool) {
	_, chain := setupChain(t, nil)
	setupTestSandboxSC(t, chain, nil, w)

	req := solo.NewCallParams(ScName, sbtestsc.FuncPanicFullEP.Name)
	_, err := chain.PostRequestSync(req, nil)
	testmisc.RequireErrorToBe(t, err, sbtestsc.MsgFullPanic)

	recStr := chain.GetRequestReceiptsForBlockRangeAsStrings(0, 0)
	str := strings.Join(recStr, "\n")
	t.Logf("\n%s", str)
	extra := 0
	if w {
		extra = 1
	}
	require.EqualValues(t, 5+extra, strings.Count(str, "Block/Request index:"))
	require.EqualValues(t, 1, strings.Count(str, sbtestsc.MsgFullPanic))
}

func TestPanicViewCall(t *testing.T) { run2(t, testPanicViewCall) }
func testPanicViewCall(t *testing.T, w bool) {
	_, chain := setupChain(t, nil)
	setupTestSandboxSC(t, chain, nil, w)

	_, err := chain.CallView(ScName, sbtestsc.FuncPanicViewEP.Name)
	testmisc.RequireErrorToBe(t, err, sbtestsc.MsgViewPanic)

	recStr := chain.GetRequestReceiptsForBlockRangeAsStrings(0, 0)
	str := strings.Join(recStr, "\n")
	t.Logf("\n%s", str)
	extra := 0
	if w {
		extra = 1
	}
	require.EqualValues(t, 4+extra, strings.Count(str, "Block/Request index:"))
	require.EqualValues(t, 0, strings.Count(str, sbtestsc.MsgViewPanic))
}

func TestCallPanicFull(t *testing.T) { run2(t, testCallPanicFull) }
func testCallPanicFull(t *testing.T, w bool) {
	_, chain := setupChain(t, nil)
	setupTestSandboxSC(t, chain, nil, w)

	req := solo.NewCallParams(ScName, sbtestsc.FuncCallPanicFullEP.Name).WithGasBudget(1000)
	_, err := chain.PostRequestSync(req, nil)
	testmisc.RequireErrorToBe(t, err, sbtestsc.MsgFullPanic)

	recStr := chain.GetRequestReceiptsForBlockRangeAsStrings(0, 0)
	str := strings.Join(recStr, "\n")
	t.Logf("\n%s", str)
	extra := 0
	if w {
		extra = 1
	}
	require.EqualValues(t, 5+extra, strings.Count(str, "Block/Request index:"))
	require.EqualValues(t, 1, strings.Count(str, sbtestsc.MsgFullPanic))
}

func TestCallPanicViewFromFull(t *testing.T) { run2(t, testCallPanicViewFromFull) }
func testCallPanicViewFromFull(t *testing.T, w bool) {
	_, chain := setupChain(t, nil)
	setupTestSandboxSC(t, chain, nil, w)

	req := solo.NewCallParams(ScName, sbtestsc.FuncCallPanicViewEPFromFull.Name).WithGasBudget(1000)
	_, err := chain.PostRequestSync(req, nil)
	testmisc.RequireErrorToBe(t, err, sbtestsc.MsgViewPanic)

	recStr := chain.GetRequestReceiptsForBlockRangeAsStrings(0, 0)
	str := strings.Join(recStr, "\n")
	t.Logf("\n%s", str)
	extra := 0
	if w {
		extra = 1
	}
	require.EqualValues(t, 5+extra, strings.Count(str, "Block/Request index:"))
	require.EqualValues(t, 1, strings.Count(str, sbtestsc.MsgViewPanic))
}

func TestCallPanicViewFromView(t *testing.T) { run2(t, testCallPanicViewFromView) }
func testCallPanicViewFromView(t *testing.T, w bool) {
	_, chain := setupChain(t, nil)
	setupTestSandboxSC(t, chain, nil, w)

	_, err := chain.CallView(ScName, sbtestsc.FuncCallPanicViewEPFromView.Name)
	testmisc.RequireErrorToBe(t, err, sbtestsc.MsgViewPanic)

	recStr := chain.GetRequestReceiptsForBlockRangeAsStrings(0, 0)
	str := strings.Join(recStr, "\n")
	t.Logf("\n%s", str)
	extra := 0
	if w {
		extra = 1
	}
	require.EqualValues(t, 4+extra, strings.Count(str, "Block/Request index:"))
	require.EqualValues(t, 0, strings.Count(str, sbtestsc.MsgViewPanic))
}
