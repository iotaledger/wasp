// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package consGR

import (
	"context"
	"fmt"
	"time"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/metrics"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/runvm"
)

type vmAsync struct {
	runner  vm.VMRunner
	metrics metrics.IChainConsensusMetrics
	log     *logger.Logger
}

func NewVMAsync(metrics metrics.IChainConsensusMetrics, log *logger.Logger) VM {
	return &vmAsync{
		runner:  runvm.NewVMRunner(),
		metrics: metrics,
		log:     log,
	}
}

// Context is not used here. There is no way to cancel a VM run.
func (vma *vmAsync) ConsensusRunTask(ctx context.Context, task *vm.VMTask) <-chan *vm.VMTask {
	respCh := make(chan *vm.VMTask, 1)
	go vma.run(task, respCh)
	return respCh
}

func (vma *vmAsync) run(task *vm.VMTask, respCh chan *vm.VMTask) {
	startTime := time.Now()
	reqCount := len(task.Requests)
	err := vma.runner.Run(task)
	runTime := time.Since(startTime)
	vma.metrics.VMRun(runTime, reqCount)
	vma.log.Debugf("VM processed %v requests in %v", reqCount, runTime)
	if err != nil {
		panic(fmt.Errorf("error running the VM:  %w", err))
	}
	respCh <- task
}
