// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package bracha

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/require"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/packages/testutil/testval"
)

func TestMsgBrachaSerialization(t *testing.T) {
	{
		b := make([]byte, 10)
		_, err := rand.Read(b)
		require.NoError(t, err)
		msg := &msgBracha{
			msgBrachaTypePropose,
			b,
		}

		bcs.TestCodec(t, msg)
	}
	{
		msg := &msgBracha{
			msgBrachaTypePropose,
			testval.TestBytes(10),
		}

		bcs.TestCodecAndHash(t, msg, "fafb2a25ad65")
	}
	{
		b := make([]byte, 10)
		_, err := rand.Read(b)
		require.NoError(t, err)
		msg := &msgBracha{
			msgBrachaTypeEcho,
			b,
		}

		bcs.TestCodec(t, msg)
	}
	{
		msg := &msgBracha{
			msgBrachaTypeEcho,
			testval.TestBytes(10),
		}

		bcs.TestCodecAndHash(t, msg, "46ca7766e199")
	}
	{
		b := make([]byte, 10)
		_, err := rand.Read(b)
		require.NoError(t, err)
		msg := &msgBracha{
			msgBrachaTypeReady,
			b,
		}

		bcs.TestCodec(t, msg)
	}
	{
		msg := &msgBracha{
			msgBrachaTypeReady,
			testval.TestBytes(10),
		}

		bcs.TestCodecAndHash(t, msg, "13fb21f67718")
	}
}
