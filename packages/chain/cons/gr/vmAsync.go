// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package consGR

import (
	"context"
	"fmt"

	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/runvm"
)

type vmAsync struct {
	runner vm.VMRunner
}

func NewVMAsync() VM {
	return &vmAsync{runner: runvm.NewVMRunner()}
}

// Context is not used here. There is no way to cancel a VM run.
func (vma *vmAsync) ConsensusRunTask(ctx context.Context, task *vm.VMTask) <-chan *vm.VMTask {
	respCh := make(chan *vm.VMTask, 1)
	go vma.run(task, respCh)
	return respCh
}

func (vma *vmAsync) run(task *vm.VMTask, respCh chan *vm.VMTask) {
	if err := vma.runner.Run(task); err != nil {
		panic(fmt.Errorf("error running the VM:  %w", err))
	}
	respCh <- task
}
