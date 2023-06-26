package chainmanager

import (
	"crypto/rand"
	"testing"

	iotago "github.com/iotaledger/iota.go/v3"

	"github.com/iotaledger/wasp/packages/util/rwutil"
)

func TestMsgCmtLogSerialization(t *testing.T) {
	// FIXME
	b := make([]byte, iotago.Ed25519AddressBytesLength)
	rand.Read(b)
	msg := &msgCmtLog{
		iotago.Ed25519Address(b),
		&msgCmtLog{},
	}

	rwutil.ReadWriteTest(t, msg, new(msgCmtLog))
}
