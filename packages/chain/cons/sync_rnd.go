// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cons

import (
	"github.com/iotaledger/wasp/packages/gpa"
)

type SyncRND interface {
	CanProceed(dataToSign []byte) gpa.OutMessages
	BLSPartialSigReceived(sender gpa.NodeID, partialSig []byte) gpa.OutMessages
}

type syncRNDImpl struct {
	blsThreshold     int
	blsPartialSigs   map[gpa.NodeID][]byte
	dataToSign       []byte
	inputsReadyCB    func(dataToSign []byte) gpa.OutMessages
	sigSharesReady   bool
	sigSharesReadyCB func(dataToSign []byte, sigShares map[gpa.NodeID][]byte) (bool, gpa.OutMessages)
}

func NewSyncRND(
	blsThreshold int,
	inputsReadyCB func(dataToSign []byte) gpa.OutMessages,
	sigSharesReadyCB func(dataToSign []byte, sigShares map[gpa.NodeID][]byte) (bool, gpa.OutMessages),
) SyncRND {
	return &syncRNDImpl{
		blsThreshold:     blsThreshold,
		blsPartialSigs:   map[gpa.NodeID][]byte{},
		inputsReadyCB:    inputsReadyCB,
		sigSharesReadyCB: sigSharesReadyCB,
	}
}

func (sub *syncRNDImpl) CanProceed(dataToSign []byte) gpa.OutMessages {
	if sub.dataToSign != nil || dataToSign == nil {
		return nil
	}
	sub.dataToSign = dataToSign
	return gpa.NoMessages().
		AddAll(sub.inputsReadyCB(sub.dataToSign)).
		AddAll(sub.tryComplete())
}

func (sub *syncRNDImpl) BLSPartialSigReceived(sender gpa.NodeID, partialSig []byte) gpa.OutMessages {
	if _, ok := sub.blsPartialSigs[sender]; ok {
		return nil // Duplicate, ignore it.
	}
	sub.blsPartialSigs[sender] = partialSig
	return sub.tryComplete()
}

func (sub *syncRNDImpl) tryComplete() gpa.OutMessages {
	if sub.sigSharesReady || sub.dataToSign == nil || len(sub.blsPartialSigs) < sub.blsThreshold {
		return nil
	}
	done, msgs := sub.sigSharesReadyCB(sub.dataToSign, sub.blsPartialSigs)
	sub.sigSharesReady = done
	return msgs
}
