// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package acss

import (
	cryptorand "crypto/rand"
	"math"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

func TestMsgImplicateRecoverSerialization(t *testing.T) {
	{
		b := make([]byte, 10)
		_, err := cryptorand.Read(b)
		require.NoError(t, err)
		msg := &msgImplicateRecover{
			gpa.NodeID{},
			gpa.NodeID{},
			msgImplicateRecoverKindIMPLICATE,
			int(uint16(rand.Intn(math.MaxUint16 + 1))),
			b,
		}

		rwutil.ReadWriteTest(t, msg, new(msgImplicateRecover))
	}
	{
		b := make([]byte, 10)
		_, err := cryptorand.Read(b)
		require.NoError(t, err)
		msg := &msgImplicateRecover{
			gpa.NodeID{},
			gpa.NodeID{},
			msgImplicateRecoverKindRECOVER,
			int(uint16(rand.Intn(math.MaxUint16 + 1))),
			b,
		}

		rwutil.ReadWriteTest(t, msg, new(msgImplicateRecover))
	}
}
