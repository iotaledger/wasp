package chainmanager

import (
	"testing"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

func TestMsgBlockProducedSerialization(t *testing.T) {
	// FIXME
	msg := &msgBlockProduced{
		gpa.BasicMessage{},
		&iotago.Transaction{},
		state.NewBlock(),
	}

	rwutil.ReadWriteTest(t, msg, new(msgBlockProduced))
}
