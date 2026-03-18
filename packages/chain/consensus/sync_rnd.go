// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package consensus

import (
	"slices"

	"github.com/iotaledger/wasp/v2/packages/gpa"
)

type SyncRND struct {
	c              *Consensus
	blsThreshold   int
	blsPartialSigs map[gpa.NodeID][]byte
	dataToSign     []byte
	sigSharesReady bool
}

func NewSyncRND(
	blsThreshold int,
	c *Consensus,
) *SyncRND {
	return &SyncRND{
		blsThreshold:   blsThreshold,
		blsPartialSigs: map[gpa.NodeID][]byte{},
		c:              c,
	}
}

func (sub *SyncRND) CanProceed(dataToSign []byte) []gpa.MessageOut {
	if sub.dataToSign != nil || dataToSign == nil {
		return nil
	}
	sub.dataToSign = dataToSign
	return slices.Concat(
		sub.c.uponRNDInputsReady(sub.dataToSign),
		sub.tryComplete(),
	)
}

func (sub *SyncRND) BLSPartialSigReceived(sender gpa.NodeID, partialSig []byte) []gpa.MessageOut {
	if _, ok := sub.blsPartialSigs[sender]; ok {
		return nil // Duplicate, ignore it.
	}
	sub.blsPartialSigs[sender] = partialSig
	return sub.tryComplete()
}

func (sub *SyncRND) tryComplete() []gpa.MessageOut {
	if sub.sigSharesReady || sub.dataToSign == nil || len(sub.blsPartialSigs) < sub.blsThreshold {
		return nil
	}
	done, msgs := sub.c.uponRNDSigSharesReady(sub.dataToSign, sub.blsPartialSigs)
	sub.sigSharesReady = done
	return msgs
}
