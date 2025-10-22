// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package consensus

import (
	"fmt"
	"strings"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/state"
)

type SyncTX struct {
	c *Consensus

	decidedAnchor *isc.StateAnchor
	unsignedTX    *iotago.TransactionData
	signature     []byte
	blockSaved    bool
	block         state.Block

	inputsReady bool
}

func NewSyncTX(c *Consensus) *SyncTX {
	return &SyncTX{c: c}
}

func (sub *SyncTX) AnchorDecided(ao *isc.StateAnchor) gpa.OutMessages {
	if sub.decidedAnchor != nil || ao == nil {
		return nil
	}
	sub.decidedAnchor = ao
	return sub.tryCompleteInputs()
}

func (sub *SyncTX) UnsignedTXReceived(unsignedTX *iotago.TransactionData) gpa.OutMessages {
	if sub.unsignedTX != nil || unsignedTX == nil {
		return nil
	}
	sub.unsignedTX = unsignedTX
	return sub.tryCompleteInputs()
}

func (sub *SyncTX) SignatureReceived(signature []byte) gpa.OutMessages {
	if sub.signature != nil || signature == nil {
		return nil
	}
	sub.signature = signature
	return sub.tryCompleteInputs()
}

func (sub *SyncTX) BlockSaved(block state.Block) gpa.OutMessages {
	if sub.blockSaved {
		return nil
	}
	sub.blockSaved = true
	sub.block = block
	return sub.tryCompleteInputs()
}

func (sub *SyncTX) tryCompleteInputs() gpa.OutMessages {
	if sub.inputsReady || sub.decidedAnchor == nil || sub.unsignedTX == nil || sub.signature == nil || !sub.blockSaved {
		return nil
	}
	sub.inputsReady = true
	return sub.c.uponTXInputsReady(sub.decidedAnchor, sub.unsignedTX, sub.block, sub.signature)
}

// String tries to provide useful human-readable compact status.
func (sub *SyncTX) String() string {
	str := "TX"
	if sub.inputsReady {
		str += statusStrOK
	} else {
		wait := []string{}
		if sub.decidedAnchor == nil {
			wait = append(wait, "decidedAnchor")
		}
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
