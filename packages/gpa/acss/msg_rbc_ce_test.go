// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package acss

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/require"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/packages/testutil/testval"
)

func TestMsgRBCCEPayloadSerialization(t *testing.T) {
	b := make([]byte, 10)
	_, err := rand.Read(b)
	require.NoError(t, err)
	msg := &MsgRBCCEPayload{
		b,
	}

	bcs.TestCodec(t, msg)

	msg = &MsgRBCCEPayload{
		testval.TestBytes(10),
	}

	bcs.TestCodecAndHash(t, msg, "9a5a2e001fcf")
}
