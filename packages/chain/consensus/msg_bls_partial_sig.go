// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package consensus

import (
	"github.com/iotaledger/wasp/v2/packages/gpa"
)

type msgBLSPartialSig struct {
	partialSig []byte `bcs:"export"`
}

func newMsgBLSPartialSig(recipient gpa.NodeID, partialSig []byte) gpa.PayloadOut {
	return gpa.NewPayloadOut(recipient, msgBLSPartialSig{
		partialSig: partialSig,
	})
}
