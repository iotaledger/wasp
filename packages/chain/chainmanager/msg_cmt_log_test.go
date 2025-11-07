package chainmanager

import (
	"math/rand"
	"testing"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/packages/chain/committeelog"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/gpa"
)

func TestMsgCommitteeLogSerialization(t *testing.T) {
	address := cryptolib.NewRandomAddress()
	msg := &msgNextLogIndex{
		*address,
		committeelog.MsgNextLogIndex{
			BasicMessage: gpa.BasicMessage{},
			NextLogIndex: committeelog.LogIndex(rand.Int31()),
			PleaseRepeat: false,
		},
	}

	bcs.TestCodec(t, msg)

	msg = &msgNextLogIndex{
		*cryptolib.TestAddress,
		committeelog.MsgNextLogIndex{
			BasicMessage: gpa.BasicMessage{},
			NextLogIndex: committeelog.LogIndex(1234567890),
			PleaseRepeat: false,
		},
	}

	bcs.TestCodecAndHash(t, msg, "27abfd74cb8e")
}
