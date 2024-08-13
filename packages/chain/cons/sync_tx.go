// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cons

import (
	"fmt"
	"strings"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/sui-go/sui"
)

type SyncTX interface {
	UnsignedTXReceived(unsignedTX *sui.TransactionData) gpa.OutMessages
	SignatureReceived(signature []byte) gpa.OutMessages
	BlockSaved(block state.Block) gpa.OutMessages
	String() string
}

type syncTXImpl struct {
	unsignedTX *sui.TransactionData
	signature  []byte
	blockSaved bool
	block      state.Block

	inputsReady   bool
	inputsReadyCB func(unsignedTX *sui.TransactionData, block state.Block, signature []byte) gpa.OutMessages
}

func NewSyncTX(inputsReadyCB func(unsignedTX *sui.TransactionData, block state.Block, signature []byte) gpa.OutMessages) SyncTX {
	return &syncTXImpl{inputsReadyCB: inputsReadyCB}
}

func (sub *syncTXImpl) UnsignedTXReceived(unsignedTX *sui.TransactionData) gpa.OutMessages {
	if sub.unsignedTX != nil || unsignedTX == nil {
		return nil
	}
	sub.unsignedTX = unsignedTX
	return sub.tryCompleteInputs()
}

func (sub *syncTXImpl) SignatureReceived(signature []byte) gpa.OutMessages {
	if sub.signature != nil || signature == nil {
		return nil
	}
	sub.signature = signature
	return sub.tryCompleteInputs()
}

func (sub *syncTXImpl) BlockSaved(block state.Block) gpa.OutMessages {
	if sub.blockSaved {
		return nil
	}
	sub.blockSaved = true
	sub.block = block
	return sub.tryCompleteInputs()
}

func (sub *syncTXImpl) tryCompleteInputs() gpa.OutMessages {
	if sub.inputsReady || sub.unsignedTX == nil || sub.signature == nil || !sub.blockSaved {
		return nil
	}
	sub.inputsReady = true
	return sub.inputsReadyCB(sub.unsignedTX, sub.block, sub.signature)
}

// Try to provide useful human-readable compact status.
func (sub *syncTXImpl) String() string {
	str := "TX"
	if sub.inputsReady {
		str += statusStrOK
	} else {
		wait := []string{}
		if sub.unsignedTX == nil {
			wait = append(wait, "unsignedTX")
		}
		if sub.signature == nil {
			wait = append(wait, "Signature")
		}
		if !sub.blockSaved {
			wait = append(wait, "SavedBlock")
		}
		str += fmt.Sprintf("/WAIT[%v]", strings.Join(wait, ","))
	}
	return str
}
