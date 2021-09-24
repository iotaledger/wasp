package runvm

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxoutil"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp/colored"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/optimism"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/vmcontext"
	"golang.org/x/xerrors"
)

type VMRunner struct{}

func (r VMRunner) Run(task *vm.VMTask) {
	// panic catcher for the whole VM task
	// ir returns gracefully if the panic was about invalidated state
	// otherwise it panics again
	// The smart contract panics are processed inside and do not reach this point
	defer func() {
		r := recover()
		if r == nil {
			return
		}
		if _, ok := r.(*optimism.ErrorStateInvalidated); ok {
			task.Log.Warnf("VM task has been abandoned due to invalidated state. ACS session id: %d", task.ACSSessionID)
			return
		}
		panic(r)
	}()
	runTask(task)
}

func NewVMRunner() VMRunner {
	return VMRunner{}
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
	// TODO access and consensus pledge
	txb := utxoutil.NewBuilder(task.ChainInput)

	vmctx, err := vmcontext.CreateVMContext(task, txb)
	if err != nil {
		task.Log.Panicf("runTask: CreateVMContext: %v", err)
	}

	var lastResult dict.Dict
	var lastErr error
	var lastTotalAssets colored.Balances
	var exceededBlockOutputLimit bool

	// loop over the batch of requests and run each request on the VM.
	// the result accumulates in the VMContext and in the list of stateUpdates
	var numOffLedger, numSuccess uint16
	var numOnLedger uint8
	for i, req := range task.Requests {
		if req.IsOffLedger() {
			numOffLedger++
		} else {
			if numOnLedger == vmcontext.MaxBlockInputCount {
				break // max number of inputs to be included in state transition reached, do not process more on-ledger requests
			}
			numOnLedger++
		}

		vmctx.RunTheRequest(req, uint16(i))
		lastResult, lastTotalAssets, lastErr, exceededBlockOutputLimit = vmctx.GetResult()

		if exceededBlockOutputLimit {
			// current request exceeded the number of output limit and will be re-run next batch
			if req.IsOffLedger() {
				numOffLedger--
			} else {
				numOnLedger--
			}
			continue
		}
		task.ProcessedRequestsCount++
		if lastErr == nil {
			numSuccess++
		} else {
			task.Log.Debugf("runTask, ERROR running request: %s, error: %v", req.ID().Base58(), lastErr)
		}
	}

	task.Log.Debugf("runTask, ran %d requests. success: %d, offledger: %d", task.ProcessedRequestsCount, numSuccess, numOffLedger)

	// save the block info into the 'blocklog' contract
	// if rotationAddr != nil ir means the block is a rotation block
	rotationAddr := vmctx.CloseVMContext(task.ProcessedRequestsCount, numSuccess, numOffLedger)

	if rotationAddr == nil {
		task.ResultTransactionEssence, err = vmctx.BuildTransactionEssence(task.VirtualState.StateCommitment(), task.VirtualState.Timestamp())
		if err != nil {
			task.OnFinish(nil, nil, xerrors.Errorf("RunVM.BuildTransactionEssence: %w", err))
			return
		}
		if err = checkTotalAssets(task.ResultTransactionEssence, lastTotalAssets); err != nil {
			task.Log.Panic(xerrors.Errorf("RunVM.checkTotalAssets: %w", err))
		}
		task.Log.Debug("runTask OUT. ",
			"block index: ", task.VirtualState.BlockIndex(),
			" variable state hash: ", task.VirtualState.StateCommitment().String(),
			" tx essence hash: ", hashing.HashData(task.ResultTransactionEssence.Bytes()).String(),
			" tx finalTimestamp: ", task.ResultTransactionEssence.Timestamp(),
		)
	} else {
		task.RotationAddress = rotationAddr
		task.ResultTransactionEssence = nil
		task.Log.Debugf("runTask OUT: rotate to address %s", rotationAddr.Base58())
	}
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
