package iotago_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
)

func TestTransactionData(t *testing.T) {
	sender := iotago.MustAddressFromHex("0x0")
	b := [32]byte{}
	b58 := iotago.Digest(b[:])
	gasPaymentsRef := iotago.ObjectRef{
		ObjectID: iotago.MustObjectIDFromHex("0x0"),
		Version:  0,
		Digest:   &b58,
	}

	ptb := iotago.NewProgrammableTransactionBuilder()
	ptb.Command(
		iotago.Command{
			MoveCall: &iotago.ProgrammableMoveCall{
				Package:       iotago.IotaPackageIDIotaSystem,
				Module:        iotago.IotaSystemModuleName,
				Function:      "function",
				TypeArguments: []iotago.TypeTag{},
				Arguments:     []iotago.Argument{},
			},
		},
	)
	pt := ptb.Finish()
	tx := iotago.NewProgrammable(
		sender,
		pt,
		[]*iotago.ObjectRef{&gasPaymentsRef},
		100000,
		1000,
	)

	targetHash := []byte{
		0x4e, 0xa0, 0xfd, 0xbf, 0xd4, 0xad, 0x50, 0x6a, 0x78, 0x34, 0x1d, 0xc0, 0x48, 0xca, 0x88, 0xa, 0xf9, 0x46,
		0xda, 0xc1, 0x46, 0xd1, 0xbb, 0x22, 0xe7, 0x86, 0x5b, 0x8a, 0x73, 0x52, 0x48, 0xb7,
	}
	digest, err := tx.Digest()
	require.NoError(t, err)
	require.Equal(t, targetHash, digest.Bytes())

	bcs.TestCodecAndHash(t, tx, "2ce05d5947a7")
}
