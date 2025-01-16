// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package distsync

import (
	"math"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/isctest"
	"github.com/iotaledger/wasp/packages/util/bcs"
)

func TestMsgShareRequestSerialization(t *testing.T) {
	{
		req := isc.NewOffLedgerRequest(isctest.RandomChainID(), isc.NewMessage(3, 14, isc.NewCallArguments([]byte{1, 2, 3})), 1337, 100).Sign(cryptolib.NewKeyPair())
		msg := &msgShareRequest{
			gpa.BasicMessage{},
			byte(rand.Intn(math.MaxUint8)),
			req,
		}

		bcs.TestCodec(t, msg)
	}
	{
		sender := cryptolib.NewRandomAddress()
		req, err := isc.OnLedgerFromRequest(isctest.RandomRequestWithRef(), sender)
		require.NoError(t, err)

		msg := &msgShareRequest{
			gpa.BasicMessage{},
			byte(rand.Intn(math.MaxUint8)),
			req,
		}

		bcs.TestCodec(t, msg)
	}
}
