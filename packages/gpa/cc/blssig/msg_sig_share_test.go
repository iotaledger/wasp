// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package blssig

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/require"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/testutil/testval"
)

func TestMsgSigShareSerialization(t *testing.T) {
	b := make([]byte, 10)
	_, err := rand.Read(b)
	require.NoError(t, err)
	msg := &msgSigShare{
		gpa.BasicMessage{},
		b,
	}
	bcs.TestCodec(t, msg)

	msg = &msgSigShare{
		gpa.BasicMessage{},
		testval.TestBytes(10),
	}
	bcs.TestCodecAndHash(t, msg, "9a5a2e001fcf")
}
