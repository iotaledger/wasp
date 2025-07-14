// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package bracha

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/require"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/testutil"
)

func TestMsgBrachaSerialization(t *testing.T) {
	{
		b := make([]byte, 10)
		_, err := rand.Read(b)
		require.NoError(t, err)
		msg := &msgBracha{
			gpa.BasicMessage{},
			msgBrachaTypePropose,
			b,
		}

		bcs.TestCodec(t, msg)
	}
	{
		msg := &msgBracha{
			gpa.BasicMessage{},
			msgBrachaTypePropose,
			testutil.TestBytes(10),
		}

		bcs.TestCodecAndHash(t, msg, "46ca7766e199")
	}
	{
		b := make([]byte, 10)
		_, err := rand.Read(b)
		require.NoError(t, err)
		msg := &msgBracha{
			gpa.BasicMessage{},
			msgBrachaTypeEcho,
			b,
		}

		bcs.TestCodec(t, msg)
	}
	{
		msg := &msgBracha{
			gpa.BasicMessage{},
			msgBrachaTypeEcho,
			testutil.TestBytes(10),
		}

		bcs.TestCodecAndHash(t, msg, "13fb21f67718")
	}
	{
		b := make([]byte, 10)
		_, err := rand.Read(b)
		require.NoError(t, err)
		msg := &msgBracha{
			gpa.BasicMessage{},
			msgBrachaTypeReady,
			b,
		}

		bcs.TestCodec(t, msg)
	}
	{
		msg := &msgBracha{
			gpa.BasicMessage{},
			msgBrachaTypeReady,
			testutil.TestBytes(10),
		}

		bcs.TestCodecAndHash(t, msg, "131d4ae6fdab")
	}
}
