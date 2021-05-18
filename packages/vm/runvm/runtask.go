package runvm

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxoutil"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/vmcontext"
	"golang.org/x/xerrors"
)

type vmRunner struct{}

func (r vmRunner) Run(task *vm.VMTask) {
	runTask(task)
}

func NewVMRunner() vmRunner {
	return vmRunner{}
}

// runTask runs batch of requests on VM
func runTask(task *vm.VMTask) {
	task.Log.Debugw("runTask IN",
		"chainID", task.ChainInput.Address().Base58(),
		"timestamp", task.Timestamp,
		"block index", task.VirtualState.BlockIndex(),
		"num req", len(task.Requests),
	)
	if len(task.Requests) == 0 {
		task.Log.Panicf("MustRunVMTaskAsync: must be at least 1 request")
	}
	outputs := outputsFromRequests(task.Requests...)
	txb := utxoutil.NewBuilder(append(outputs, task.ChainInput)...)

	vmctx, err := vmcontext.CreateVMContext(task, txb)
	if err != nil {
		task.Log.Panicf("runTask: CreateVMContext: %v", err)
	}

	var lastResult dict.Dict
	var lastErr error
	var lastTotalAssets *ledgerstate.ColoredBalances

	// loop over the batch of requests and run each request on the VM.
	// the result accumulates in the VMContext and in the list of stateUpdates
	var numOffLedger, numSuccess uint16
	for i, req := range task.Requests {
		vmctx.RunTheRequest(req, uint16(i))
		lastResult, lastTotalAssets, lastErr = vmctx.GetResult()

		if req.Output() == nil {
			numOffLedger++
		}
		if lastErr == nil {
			numSuccess++
		}
	}

	// save the block info into the 'blocklog' contract
	err = vmctx.CloseVMContext(uint16(len(task.Requests)), numSuccess, numOffLedger)
	if err != nil {
		task.OnFinish(nil, nil, xerrors.Errorf("RunVM: %w", err))
		return
	}

	task.ResultTransactionEssence, err = vmctx.BuildTransactionEssence(task.VirtualState.Hash(), task.VirtualState.Timestamp())
	if err != nil {
		task.OnFinish(nil, nil, xerrors.Errorf("RunVM.BuildTransactionEssence: %v", err))
		return
	}
	if err = checkTotalAssets(task.ResultTransactionEssence, lastTotalAssets); err != nil {
		task.OnFinish(nil, nil, xerrors.Errorf("RunVM.checkTotalAssets: %v", err))
		return
	}
	task.Log.Debugw("runTask OUT",
		"block index", task.VirtualState.BlockIndex(),
		"variable state hash", task.VirtualState.Hash().Bytes(),
		"tx essence hash", hashing.HashData(task.ResultTransactionEssence.Bytes()).String(),
		"tx finalTimestamp", task.ResultTransactionEssence.Timestamp(),
	)
	task.OnFinish(lastResult, lastErr, nil)
}

// outputsFromRequests collect all outputs from requests which are on-ledger
func outputsFromRequests(requests ...coretypes.Request) []ledgerstate.Output {
	ret := make([]ledgerstate.Output, 0)
	for _, req := range requests {
		if out := req.Output(); out != nil {
			ret = append(ret, out)
		}
	}
	return ret
}

func checkTotalAssets(essence *ledgerstate.TransactionEssence, lastTotalAssets *ledgerstate.ColoredBalances) error {
	var chainOutput *ledgerstate.AliasOutput
	for _, o := range essence.Outputs() {
		if out, ok := o.(*ledgerstate.AliasOutput); ok {
			chainOutput = out
		}
	}
	if chainOutput == nil {
		return xerrors.New("inconsistency: chain output not found")
	}
	diffAssets := util.DiffColoredBalances(chainOutput.Balances(), lastTotalAssets)
	if iotas, ok := diffAssets[ledgerstate.ColorIOTA]; !ok || iotas != ledgerstate.DustThresholdAliasOutputIOTA {
		return xerrors.Errorf("RunVM.BuildTransactionEssence: inconsistency between L1 and L2 ledgers")
	}
	return nil
}
