// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package gr

import (
	"context"
	"fmt"
	"time"

	"github.com/iotaledger/hive.go/log"

	"github.com/iotaledger/wasp/packages/metrics"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/vmimpl"
)

type vmAsync struct {
	metrics *metrics.ChainConsensusMetrics
	log     log.Logger
}

func NewVMAsync(metrics *metrics.ChainConsensusMetrics, log log.Logger) VM {
	return &vmAsync{
		metrics: metrics,
		log:     log,
	}
}

// Context is not used here. There is no way to cancel a VM run.
func (vma *vmAsync) ConsensusRunTask(ctx context.Context, task *vm.VMTask) <-chan *vm.VMTaskResult {
	respCh := make(chan *vm.VMTaskResult, 1)
	go vma.run(task, respCh)
	return respCh
}

func (vma *vmAsync) run(task *vm.VMTask, respCh chan *vm.VMTaskResult) {
	startTime := time.Now()
	reqCount := len(task.Requests)
	vmResult, err := vmimpl.Run(task)
	runTime := time.Since(startTime)
	vma.metrics.VMRun(runTime, reqCount)
	vma.log.LogDebugf("VM processed %v requests in %v", reqCount, runTime)
	if err != nil {
		panic(fmt.Errorf("error running the VM:  %w", err))
	}
	respCh <- vmResult
}
