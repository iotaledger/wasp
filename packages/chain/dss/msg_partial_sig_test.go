// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package dss

import (
	"testing"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

func TestMsgPartialSigSerialization(t *testing.T) {
	// FIXME
	msg := &msgPartialSig{
		gpa.BasicMessage{},
		tcrypto.DefaultEd25519Suite(),
		nil,
	}

	rwutil.ReadWriteTest(t, msg, new(msgPartialSig))
}
