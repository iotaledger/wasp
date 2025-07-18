package iotasigner_test

import (
	"testing"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago/serialization"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotasigner"
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
