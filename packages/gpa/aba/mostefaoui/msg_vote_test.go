package mostefaoui

import (
	"math"
	"testing"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/util/bcs"
)

func TestMsgVoteCodec(t *testing.T) {
	msg := &msgVote{
		gpa.BasicMessage{},
		math.MaxUint16,
		AUX,
		true,
	}

	bcs.TestCodec(t, msg)
}
