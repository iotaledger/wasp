package mostefaoui

import (
	"math"
	"testing"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/packages/gpa"
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
