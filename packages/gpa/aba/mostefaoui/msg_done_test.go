// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package mostefaoui

import (
	"math"
	"math/rand"
	"testing"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/util/bcs"
)

func TestMsgDoneSerialization(t *testing.T) {
	msg := &msgDone{
		gpa.BasicMessage{},
		int(uint16(rand.Intn(math.MaxUint16 + 1))),
	}

	bcs.TestCodec(t, msg)

	msg = &msgDone{
		gpa.BasicMessage{},
		math.MaxUint16,
	}

	bcs.TestCodec(t, msg)
}
