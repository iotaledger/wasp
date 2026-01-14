// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package mostefaoui

import (
	"math"
	"math/rand"
	"testing"

	bcs "github.com/iotaledger/bcs-go"
)

func TestMsgDoneSerialization(t *testing.T) {
	msg := &MsgDone{
		int(uint16(rand.Intn(math.MaxUint16 + 1))),
	}

	bcs.TestCodec(t, msg)

	msg = &MsgDone{
		math.MaxUint16,
	}

	bcs.TestCodecAndHash(t, msg, "ab2affdd99ab")
}
