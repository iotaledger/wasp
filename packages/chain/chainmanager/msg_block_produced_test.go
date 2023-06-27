package chainmanager

import (
	"testing"

	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

func TestMsgBlockProducedSerialization(t *testing.T) {
	msg := &msgBlockProduced{
		gpa.BasicMessage{},
		tpkg.RandTransaction(),
		state.RandomBlock(),
	}

	rwutil.ReadWriteTest(t, msg, new(msgBlockProduced))
}
