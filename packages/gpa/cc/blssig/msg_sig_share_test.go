// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package blssig

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

func TestMsgSigShareSerialization(t *testing.T) {
	b := make([]byte, 10)
	_, err := rand.Read(b)
	require.NoError(t, err)
	msg := &msgSigShare{
		gpa.BasicMessage{},
		b,
	}
	rwutil.ReadWriteTest(t, msg, new(msgSigShare))
}
