package chainmanager

import (
	"testing"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util/rwutil"
	"github.com/iotaledger/wasp/sui-go/suisigner/suisignertest"
)

func TestMsgBlockProducedSerialization(t *testing.T) {
	randomSignedTransaction := suisignertest.RandomSignedTransaction()
	msg := &msgBlockProduced{
		gpa.BasicMessage{},
		&randomSignedTransaction,
		state.RandomBlock(),
	}

	rwutil.ReadWriteTest(t, msg, new(msgBlockProduced))
}
