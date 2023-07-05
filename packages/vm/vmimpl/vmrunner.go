package vmimpl

import (
	"github.com/iotaledger/wasp/packages/util/panicutil"
	"github.com/iotaledger/wasp/packages/vm"
)

type VMRunner struct{}

func (r VMRunner) Run(task *vm.VMTask) (res *vm.VMTaskResult, err error) {
	// top exception catcher for all panics
	// The VM session will be abandoned peacefully
	err = panicutil.CatchAllButDBError(func() {
		res = task.CreateResult()
		createVMContext(task, res).run()
	}, task.Log)
	if err != nil {
		task.Log.Warnf("GENERAL VM EXCEPTION: the task has been abandoned due to: %s", err.Error())
	}
	return res, err
}

func NewVMRunner() vm.VMRunner {
	return VMRunner{}
}
