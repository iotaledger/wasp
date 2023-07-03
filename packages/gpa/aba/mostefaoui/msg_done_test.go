// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package mostefaoui

import (
	"math"
	"math/rand"
	"testing"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

func TestMsgDoneSerialization(t *testing.T) {
	msg := &msgDone{
		gpa.BasicMessage{},
		int(uint16(rand.Intn(math.MaxUint16 + 1))),
	}

	rwutil.ReadWriteTest(t, msg, new(msgDone))
}
