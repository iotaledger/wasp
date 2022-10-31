// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cons

import (
	"fmt"
	"strings"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/vm"
)

type SyncTX interface {
	VMResultReceived(vmResult *vm.VMTask) gpa.OutMessages
	SignatureReceived(signature []byte) gpa.OutMessages
	String() string
}

type syncTXImpl struct {
	VMResult  *vm.VMTask
	Signature []byte

	inputsReady   bool
	inputsReadyCB func(vmResult *vm.VMTask, signature []byte) gpa.OutMessages
}

func NewSyncTX(inputsReadyCB func(vmResult *vm.VMTask, signature []byte) gpa.OutMessages) SyncTX {
	return &syncTXImpl{inputsReadyCB: inputsReadyCB}
}

func (sub *syncTXImpl) VMResultReceived(vmResult *vm.VMTask) gpa.OutMessages {
	if sub.VMResult != nil || vmResult == nil {
		return nil
	}
	sub.VMResult = vmResult
	return sub.tryCompleteInputs()
}

func (sub *syncTXImpl) SignatureReceived(signature []byte) gpa.OutMessages {
	if sub.Signature != nil || signature == nil {
		return nil
	}
	sub.Signature = signature
	return sub.tryCompleteInputs()
}

func (sub *syncTXImpl) tryCompleteInputs() gpa.OutMessages {
	if sub.inputsReady || sub.VMResult == nil || sub.Signature == nil {
		return nil
	}
	sub.inputsReady = true
	return sub.inputsReadyCB(sub.VMResult, sub.Signature)
}

// Try to provide useful human-readable compact status.
func (sub *syncTXImpl) String() string {
	str := "TX"
	if sub.inputsReady {
		str += "/OK"
	} else {
		wait := []string{}
		if sub.VMResult == nil {
			wait = append(wait, "VMResult")
		}
		if sub.Signature == nil {
			wait = append(wait, "Signature")
		}
		str += fmt.Sprintf("/WAIT[%v]", strings.Join(wait, ","))
	}
	return str
}
