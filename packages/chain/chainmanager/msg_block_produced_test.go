package chainmanager

import (
	"testing"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/clients/iota-go/iotasigner/iotasignertest"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/state"
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

	msg = &msgBlockProduced{
		gpa.BasicMessage{},
		&iotasignertest.TestSignedTransaction,
		state.TestBlock,
	}

	bcs.TestCodecAndHash(t, msg, "6b906810f98b", &msgBlockProduced{
		block: state.NewBlock(),
	})
}
