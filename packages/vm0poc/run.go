package vm0poc

import (
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/util/panicutil"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/errors/coreerrors"
)

type VMRunner struct{}

func NewVMRunner() VMRunner {
	return VMRunner{}
}

func (r VMRunner) Run(task *vm.VMTask) {
	// optimistic read panic catcher for the whole VM task
	err := panicutil.CatchPanicReturnError(
		func() { runTask(task) },
		coreutil.ErrorStateInvalidated,
	)
	if err != nil {
		switch e := err.(type) {
		case *iscp.VMError:
			task.VMError = e
		case error:
			// May require a different error type here?
			task.VMError = coreerrors.ErrUntypedError.Create(e.Error())
		default:
			task.VMError = coreerrors.ErrUntypedError.Create(e.Error())
		}
		task.Log.Warnf("VM task has been abandoned due to invalidated state. ACS session id: %d", task.ACSSessionID)
	}
}

func runTask(task *vm.VMTask) {
	panic("implement me")
}
