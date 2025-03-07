// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package bracha

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/require"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/packages/gpa"
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
}
