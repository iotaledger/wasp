package iotasigner_test

import (
	"testing"

	"github.com/iotaledger/wasp/clients/iota-go/iotago/serialization"
	"github.com/iotaledger/wasp/clients/iota-go/iotasigner"
	"github.com/iotaledger/wasp/packages/util/bcs"
)

func TestIntentCodec(t *testing.T) {
	bcs.TestCodecAndBytes(
		t, &iotasigner.Intent{
			Scope: iotasigner.IntentScope{
				CheckpointSummary: &serialization.EmptyEnum{},
			},
			Version: iotasigner.IntentVersion{
				V0: &serialization.EmptyEnum{},
			},
			AppID: iotasigner.AppID{
				Narwhal: &serialization.EmptyEnum{},
			},
		}, []byte{0x2, 0x0, 0x1},
	)
}
