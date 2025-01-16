package chainmanager

import (
	"math/rand"
	"testing"

	"github.com/iotaledger/wasp/packages/chain/cmt_log"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/util/bcs"
)

func TestMsgCmtLogSerialization(t *testing.T) {
	address := cryptolib.NewRandomAddress()
	msg := &msgCmtLog{
		*address,
		&cmt_log.MsgNextLogIndex{
			BasicMessage: gpa.BasicMessage{},
			NextLogIndex: cmt_log.LogIndex(rand.Int31()),
			PleaseRepeat: false,
		},
	}

	bcs.TestCodec(t, msg)
}
