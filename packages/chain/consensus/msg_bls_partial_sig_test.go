// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package consensus

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/require"

	bcs "github.com/iotaledger/bcs-go"
)

func TestMsgBLSPartialSigSerialization(t *testing.T) {
	b := make([]byte, 10)
	_, err := rand.Read(b)
	require.NoError(t, err)
	msg := &msgBLSPartialSig{
		b,
	}
	bcs.TestCodec(t, msg)

	msg = &msgBLSPartialSig{
		[]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
	}
	bcs.TestCodecAndHash(t, msg, "9a5a2e001fcf")
}
