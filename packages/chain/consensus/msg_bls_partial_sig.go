// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package consensus

import (
	"go.dedis.ch/kyber/v3/suites"

	"github.com/iotaledger/wasp/v2/packages/gpa"
)

type msgBLSPartialSig struct {
	blsSuite   suites.Suite
	partialSig []byte `bcs:"export"`
}

func newMsgBLSPartialSig(blsSuite suites.Suite, recipient gpa.NodeID, partialSig []byte) gpa.PayloadOut {
	return gpa.NewPayloadOut(recipient, &msgBLSPartialSig{
		blsSuite:   blsSuite,
		partialSig: partialSig,
	})
}
