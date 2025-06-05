package chainmanager

import (
	"math/rand"
	"testing"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/packages/chain/cmtlog"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/gpa"
)

func TestMsgCmtLogSerialization(t *testing.T) {
	address := cryptolib.NewRandomAddress()
	msg := &msgCmtLog{
		*address,
		&cmtlog.MsgNextLogIndex{
			BasicMessage: gpa.BasicMessage{},
			NextLogIndex: cmtlog.LogIndex(rand.Int31()),
			PleaseRepeat: false,
		},
	}

	bcs.TestCodec(t, msg)
}
