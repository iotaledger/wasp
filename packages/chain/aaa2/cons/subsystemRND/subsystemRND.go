// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package subsystemRND

import (
	"github.com/iotaledger/wasp/packages/gpa"
)

type SubsystemRND struct {
	blsThreshold     int
	blsPartialSigs   map[gpa.NodeID][]byte
	dataToSign       []byte
	inputsReadyCB    func(dataToSign []byte) gpa.OutMessages
	sigSharesReady   bool
	sigSharesReadyCB func(dataToSign []byte, sigShares map[gpa.NodeID][]byte) (bool, gpa.OutMessages)
}

func New(
	blsThreshold int,
	inputsReadyCB func(dataToSign []byte) gpa.OutMessages,
	sigSharesReadyCB func(dataToSign []byte, sigShares map[gpa.NodeID][]byte) (bool, gpa.OutMessages),
) *SubsystemRND {
	return &SubsystemRND{
		blsThreshold:     blsThreshold,
		blsPartialSigs:   map[gpa.NodeID][]byte{},
		inputsReadyCB:    inputsReadyCB,
		sigSharesReadyCB: sigSharesReadyCB,
	}
}

func (sub *SubsystemRND) CanProceed(dataToSign []byte) gpa.OutMessages {
	if sub.dataToSign != nil || dataToSign == nil {
		return nil
	}
	sub.dataToSign = dataToSign
	return gpa.NoMessages().
		AddAll(sub.inputsReadyCB(sub.dataToSign)).
		AddAll(sub.tryComplete())
}

func (sub *SubsystemRND) BLSPartialSigReceived(sender gpa.NodeID, partialSig []byte) gpa.OutMessages {
	if _, ok := sub.blsPartialSigs[sender]; ok {
		return nil // Duplicate, ignore it.
	}
	sub.blsPartialSigs[sender] = partialSig
	return sub.tryComplete()
}

func (sub *SubsystemRND) tryComplete() gpa.OutMessages {
	if sub.sigSharesReady || sub.dataToSign == nil || len(sub.blsPartialSigs) < sub.blsThreshold {
		return nil
	}
	done, msgs := sub.sigSharesReadyCB(sub.dataToSign, sub.blsPartialSigs)
	sub.sigSharesReady = done
	return msgs
}
