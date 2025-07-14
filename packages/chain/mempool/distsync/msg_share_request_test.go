// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package distsync

import (
	"math"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/isctest"
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
		req := isc.NewOffLedgerRequest(isctest.TestChainID, isc.NewMessage(3, 14, isc.NewCallArguments([]byte{1, 2, 3})), 1337, 100).Sign(cryptolib.TestKeyPair)
		msg := &msgShareRequest{
			gpa.BasicMessage{},
			123,
			req,
		}

		bcs.TestCodecAndHash(t, msg, "efa672554b1a")
	}
	{
		sender := cryptolib.NewRandomAddress()
		req, err := isc.OnLedgerFromMoveRequest(isctest.RandomRequestWithRef(), sender)
		require.NoError(t, err)

		msg := &msgShareRequest{
			gpa.BasicMessage{},
			byte(rand.Intn(math.MaxUint8)),
			req,
		}

		bcs.TestCodec(t, msg)
	}
	{
		sender := cryptolib.TestAddress
		req, err := isc.OnLedgerFromMoveRequest(isctest.TestRequestWithRef, sender)
		require.NoError(t, err)

		msg := &msgShareRequest{
			gpa.BasicMessage{},
			123,
			req,
		}

		bcs.TestCodecAndHash(t, msg, "82a90c016194")
	}
}
