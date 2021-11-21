package runvm

import (
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/util"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp/colored"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/vmcontext"
	"golang.org/x/xerrors"
)

type VMRunner struct{}

func (r VMRunner) Run(task *vm.VMTask) {
	// optimistic read panic catcher for the whole VM task
	err := util.CatchPanicReturnError(func() {
		runTask(task)
	}, coreutil.ErrorStateInvalidated)
	if err != nil {
		task.Log.Warnf("VM task has been abandoned due to invalidated state. ACS session id: %d", task.ACSSessionID)
	}
}

func NewVMRunner() VMRunner {
	return VMRunner{}
}

// runTask runs batch of requests on VM
func runTask(task *vm.VMTask) {
	vmctx := vmcontext.CreateVMContext(task)

	var lastResult dict.Dict
	var lastErr error
	var lastTotalAssets colored.Balances

	var numOffLedger, numSuccess uint16
	var numOnLedger uint8
	reqIndexInTheBlock := 0

	// main loop over the batch of requests and run each request on the VM.
	// the result accumulates in the VMContext and in the list of stateUpdates
	for _, req := range task.Requests {
		if skipReason := vmctx.RunTheRequest(req, uint16(reqIndexInTheBlock)); skipReason != nil {
			vmctx.Log().Warnf("request skipped (ignored) by the VM: %s, reason: %v",
				req.Request().ID().String(), skipReason)
			continue
		}
		reqIndexInTheBlock++
		if req.Type() == iscp.TypeOffLedger {
			numOffLedger++
		} else {
			numOnLedger++
		}
		lastResult, lastTotalAssets, lastErr = vmctx.GetResult()

		task.ProcessedRequestsCount++
		if lastErr == nil {
			numSuccess++
		} else {
			task.Log.Debugf("runTask, ERROR running request: %s, error: %v", req.Request().ID().String(), lastErr)
		}
	}

	task.Log.Debugf("runTask, ran %d requests. success: %d, offledger: %d", task.ProcessedRequestsCount, numSuccess, numOffLedger)

	blockIndex, stateCommitment, timestamp, rotationAddr := vmctx.CloseVMContext(task.ProcessedRequestsCount, numSuccess, numOffLedger)

	task.Log.Debugf("closed VMContext: block index: %d, state hash: %s timestamp: %v, rotationAddr: %v",
		blockIndex, stateCommitment, timestamp, rotationAddr)

	if rotationAddr == nil {
		var err error
		// rotation does not happen
		task.ResultTransactionEssence, err = vmctx.BuildTransactionEssence(stateCommitment, timestamp)
		if err != nil {
			task.OnFinish(nil, nil, xerrors.Errorf("RunVM.BuildTransactionEssence: %w", err))
			return
		}
		if err = checkTotalAssets(task.ResultTransactionEssence, lastTotalAssets); err != nil {
			task.Log.Panic(xerrors.Errorf("RunVM.checkTotalAssets: %w", err))
		}
		task.Log.Debug("runTask OUT. ",
			"block index: ", blockIndex,
			" variable state hash: ", stateCommitment,
			" tx essence hash: ", hashing.HashData(task.ResultTransactionEssence.Bytes()).String(),
			" tx timestamp: ", task.ResultTransactionEssence.Timestamp(),
		)
	} else {
		// rotation happens
		task.RotationAddress = rotationAddr
		task.ResultTransactionEssence = nil
		task.Log.Debugf("runTask OUT: rotate to address %s", rotationAddr.Base58())
	}
	// call callback closure
	task.OnFinish(lastResult, lastErr, nil)
}

func checkTotalAssets(essence *ledgerstate.TransactionEssence, lastTotalOnChainAssets colored.Balances) error {
	var chainOutput *ledgerstate.AliasOutput
	for _, o := range essence.Outputs() {
		if out, ok := o.(*ledgerstate.AliasOutput); ok {
			chainOutput = out
		}
	}
	if chainOutput == nil {
		return xerrors.New("inconsistency: chain output not found")
	}
	balancesOnOutput := colored.BalancesFromL1Balances(chainOutput.Balances())
	diffAssets := balancesOnOutput.Diff(lastTotalOnChainAssets)
	// we expect assets in the chain output and total assets on-chain differs only in the amount of
	// anti-dust tokens locked in the output. Otherwise it is inconsistency
	if len(diffAssets) != 1 || diffAssets[colored.IOTA] != int64(ledgerstate.DustThresholdAliasOutputIOTA) {
		return xerrors.Errorf("inconsistency between L1 and L2 ledgers. Diff: %+v", diffAssets)
	}
	return nil
}
