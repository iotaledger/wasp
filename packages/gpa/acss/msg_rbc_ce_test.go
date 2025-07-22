// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package acss

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/require"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/testutil"
)

func TestMsgRBCCEPayloadSerialization(t *testing.T) {
	b := make([]byte, 10)
	_, err := rand.Read(b)
	require.NoError(t, err)
	msg := &msgRBCCEPayload{
		gpa.BasicMessage{},
		nil,
		b,
		nil,
	}

	bcs.TestCodec(t, msg)

	msg = &msgRBCCEPayload{
		gpa.BasicMessage{},
		nil,
		testutil.TestBytes(10),
		nil,
	}

	bcs.TestCodecAndHash(t, msg, "9a5a2e001fcf")
}
