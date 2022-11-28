// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cons

import (
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/vm"
)

type inputVMResult struct {
	task *vm.VMTask // With results set.
}

func NewInputVMResult(task *vm.VMTask) gpa.Input {
	return &inputVMResult{task: task}
}
