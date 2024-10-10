package suisigner_test

import (
	"testing"

	"github.com/iotaledger/wasp/clients/iota-go/sui/serialization"
	"github.com/iotaledger/wasp/clients/iota-go/suisigner"
	"github.com/iotaledger/wasp/packages/util/bcs"
)

func TestIntentCodec(t *testing.T) {
	bcs.TestCodecAndBytes(t, &suisigner.Intent{
		Scope: suisigner.IntentScope{
			CheckpointSummary: &serialization.EmptyEnum{},
		},
		Version: suisigner.IntentVersion{
			V0: &serialization.EmptyEnum{},
		},
		AppID: suisigner.AppID{
			Narwhal: &serialization.EmptyEnum{},
		},
	}, []byte{0x2, 0x0, 0x1})
}
