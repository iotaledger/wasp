package sbtests

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/testutil/testmisc"
	"github.com/iotaledger/wasp/packages/vm/core/errors/coreerrors"
	"github.com/iotaledger/wasp/packages/vm/core/testcore/sbtests/sbtestsc"
)

func TestPanicFull(t *testing.T) {
	run2(t, testPanicFull)
}

func testPanicCallView(t *testing.T, w bool, funName string, gasBudget uint64, panicView string) {
	_, chain := setupChain(t, nil)
	setupTestSandboxSC(t, chain, nil, w)

	req := solo.NewCallParams(ScName, funName).WithGasBudget(gasBudget)
	_, err := chain.PostRequestSync(req, nil)
	testmisc.RequireErrorToBe(t, err, panicView)

	receipts := chain.GetRequestReceiptsForBlockRange(0, 0)
	receiptsAsString := chain.GetRequestReceiptsForBlockRangeAsStrings(0, 0)
	errorsAsString := ""

	for _, a := range receipts {
		receiptError := chain.ResolveVMError(a.Error)
		if receiptError != nil {
			errorsAsString += receiptError.Error()
		}
	}

	receiptOutput := strings.Join(receiptsAsString, "\n")
	t.Logf("\n%s", receiptOutput)
	extra := 0
	if w {
		extra = 1
	}

	require.EqualValues(t, 5+extra, strings.Count(receiptOutput, "Block/Request index:"))
	require.EqualValues(t, 1, strings.Count(errorsAsString, panicView))
}

func testPanicFull(t *testing.T, w bool) {
	testPanicCallView(t, w, sbtestsc.FuncPanicFullEP.Name, 100_000, sbtestsc.MsgFullPanic)
}

func TestPanicViewCall(t *testing.T) {
	run2(t, testPanicViewCall)
}

func testPanicViewCall(t *testing.T, w bool) {
	_, chain := setupChain(t, nil)
	setupTestSandboxSC(t, chain, nil, w)

	_, err := chain.CallView(ScName, sbtestsc.FuncPanicViewEP.Name)
	testmisc.RequireErrorToBe(t, err, sbtestsc.MsgViewPanic)

	receipts := chain.GetRequestReceiptsForBlockRange(0, 0)
	receiptsAsString := chain.GetRequestReceiptsForBlockRangeAsStrings(0, 0)
	errorsAsString := ""

	for _, a := range receipts {
		receiptError := chain.ResolveVMError(a.Error)
		if receiptError != nil {
			errorsAsString += receiptError.Error()
		}
	}

	receiptOutput := strings.Join(receiptsAsString, "\n")

	t.Logf("\n%s", receiptOutput)
	extra := 0
	if w {
		extra = 1
	}
	require.EqualValues(t, 4+extra, strings.Count(receiptOutput, "Block/Request index:"))
	require.EqualValues(t, 0, strings.Count(errorsAsString, sbtestsc.MsgViewPanic))
}

func TestCallPanicFull(t *testing.T) {
	run2(t, testCallPanicFull)
}

func testCallPanicFull(t *testing.T, w bool) {
	_, chain := setupChain(t, nil)
	setupTestSandboxSC(t, chain, nil, w)

	req := solo.NewCallParams(ScName, sbtestsc.FuncCallPanicFullEP.Name).WithGasBudget(17000)
	_, err := chain.PostRequestSync(req, nil)
	testmisc.RequireErrorToBe(t, err, coreerrors.ErrUntypedError.Create(sbtestsc.MsgFullPanic))

	receipts := chain.GetRequestReceiptsForBlockRange(0, 0)
	receiptsAsString := chain.GetRequestReceiptsForBlockRangeAsStrings(0, 0)
	errorsAsString := ""

	for _, a := range receipts {
		receiptError := chain.ResolveVMError(a.Error)
		if receiptError != nil {
			errorsAsString += receiptError.Error()
		}
	}

	receiptOutput := strings.Join(receiptsAsString, "\n")

	t.Logf("\n%s", receiptOutput)
	extra := 0
	if w {
		extra = 1
	}
	require.EqualValues(t, 5+extra, strings.Count(receiptOutput, "Block/Request index:"))
	require.EqualValues(t, 1, strings.Count(errorsAsString, sbtestsc.MsgFullPanic))
}

func TestCallPanicViewFromFull(t *testing.T) {
	run2(t, testCallPanicViewFromFull)
}

func testCallPanicViewFromFull(t *testing.T, w bool) {
	testPanicCallView(t, w, sbtestsc.FuncCallPanicViewEPFromFull.Name, 20_000, sbtestsc.MsgViewPanic)
}

func TestCallPanicViewFromView(t *testing.T) {
	run2(t, testCallPanicViewFromView)
}

func testCallPanicViewFromView(t *testing.T, w bool) {
	_, chain := setupChain(t, nil)
	setupTestSandboxSC(t, chain, nil, w)

	_, err := chain.CallView(ScName, sbtestsc.FuncCallPanicViewEPFromView.Name)
	testmisc.RequireErrorToBe(t, err, sbtestsc.MsgViewPanic)

	receipts := chain.GetRequestReceiptsForBlockRange(0, 0)
	receiptsAsString := chain.GetRequestReceiptsForBlockRangeAsStrings(0, 0)
	errorsAsString := ""

	for _, a := range receipts {
		receiptError := chain.ResolveVMError(a.Error)
		if receiptError != nil {
			errorsAsString += receiptError.Error()
		}
	}

	receiptOutput := strings.Join(receiptsAsString, "\n")

	t.Logf("\n%s", receiptOutput)
	extra := 0
	if w {
		extra = 1
	}
	require.EqualValues(t, 4+extra, strings.Count(receiptOutput, "Block/Request index:"))
	require.EqualValues(t, 0, strings.Count(errorsAsString, sbtestsc.MsgViewPanic))
}
