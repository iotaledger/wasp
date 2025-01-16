package chainmanager

import (
	"testing"

	"github.com/iotaledger/wasp/clients/iota-go/iotasigner/iotasignertest"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util/bcs"
)

func TestMsgBlockProducedSerialization(t *testing.T) {
	randomSignedTransaction := iotasignertest.RandomSignedTransaction()
	msg := &msgBlockProduced{
		gpa.BasicMessage{},
		&randomSignedTransaction,
		state.RandomBlock(),
	}

	bcs.TestCodec(t, msg, &msgBlockProduced{
		block: state.NewBlock(),
	})
}
