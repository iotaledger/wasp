// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package acss

import (
	"crypto/rand"
	"testing"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/util/rwutil"
	"github.com/stretchr/testify/require"
)

func TestMsgRBCCEPayloadSerialization(t *testing.T) {
	// FIXME
	b := make([]byte, 10)
	_, err := rand.Read(b)
	require.NoError(t, err)
	msg := &msgRBCCEPayload{
		gpa.BasicMessage{},
		nil,
		b,
	}

	rwutil.ReadWriteTest(t, msg, new(msgRBCCEPayload))
}
