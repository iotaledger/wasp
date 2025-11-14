package mostefaoui

import (
	"math"
	"testing"

	bcs "github.com/iotaledger/bcs-go"
)

func TestMsgVoteCodec(t *testing.T) {
	msg := &MsgVote{
		math.MaxUint16,
		AUX,
		true,
	}

	bcs.TestCodecAndHash(t, msg, "d54055d73f91")
}
