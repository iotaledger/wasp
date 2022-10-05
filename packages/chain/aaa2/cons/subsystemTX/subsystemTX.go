// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package subsystemTX

import (
	"fmt"
	"strings"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/vm"
)

type SubsystemTX struct {
	VMResult  *vm.VMTask
	Signature []byte

	inputsReady   bool
	inputsReadyCB func(sub *SubsystemTX) gpa.OutMessages
}

func New(inputsReadyCB func(sub *SubsystemTX) gpa.OutMessages) *SubsystemTX {
	return &SubsystemTX{inputsReadyCB: inputsReadyCB}
}

func (sub *SubsystemTX) VMResultReceived(vmResult *vm.VMTask) gpa.OutMessages {
	if sub.VMResult != nil || vmResult == nil {
		return nil
	}
	sub.VMResult = vmResult
	return sub.tryCompleteInputs()
}

func (sub *SubsystemTX) SignatureReceived(signature []byte) gpa.OutMessages {
	if sub.Signature != nil || signature == nil {
		return nil
	}
	sub.Signature = signature
	return sub.tryCompleteInputs()
}

func (sub *SubsystemTX) tryCompleteInputs() gpa.OutMessages {
	if sub.inputsReady || sub.VMResult == nil || sub.Signature == nil {
		return nil
	}
	sub.inputsReady = true
	return sub.inputsReadyCB(sub)
}

// Try to provide useful human-readable compact status.
func (sub *SubsystemTX) String() string {
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
