package chainmanager

import (
	cryptorand "crypto/rand"
	"math/rand"
	"testing"

	iotago "github.com/iotaledger/iota.go/v3"

	"github.com/iotaledger/wasp/packages/chain/cmt_log"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

func TestMsgCmtLogSerialization(t *testing.T) {
	b := make([]byte, iotago.Ed25519AddressBytesLength)
	cryptorand.Read(b)
	msg := &msgCmtLog{
		iotago.Ed25519Address(b),
		&cmt_log.MsgNextLogIndex{
			BasicMessage: gpa.BasicMessage{},
			NextLogIndex: cmt_log.LogIndex(rand.Int31()),
			PleaseRepeat: false,
		},
	}

	rwutil.ReadWriteTest(t, msg, new(msgCmtLog))
}
