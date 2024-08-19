package sui_test

import (
	"testing"

	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/stretchr/testify/require"
)

func TestTransactionData(t *testing.T) {
	sender := sui.MustAddressFromHex("0x0")
	b := [32]byte{}
	b58 := sui.Base58(b[:])
	gasPaymentsRef := sui.ObjectRef{
		ObjectID: sui.MustObjectIDFromHex("0x0"),
		Version:  0,
		Digest:   &b58,
	}

	ptb := sui.NewProgrammableTransactionBuilder()
	ptb.Command(
		sui.Command{
			MoveCall: &sui.ProgrammableMoveCall{
				Package:       sui.SuiPackageIdSuiSystem,
				Module:        sui.SuiSystemModuleName,
				Function:      "function",
				TypeArguments: []sui.TypeTag{},
				Arguments:     []sui.Argument{},
			},
		},
	)
	pt := ptb.Finish()
	tx := sui.NewProgrammable(
		sender,
		pt,
		[]*sui.ObjectRef{&gasPaymentsRef},
		100000,
		1000,
	)

	targetHash := []byte{15, 146, 17, 96, 201, 139, 191, 224, 240, 126, 40, 191, 51, 174, 182, 229, 212, 15, 130, 245, 34, 169, 114, 103, 99, 194, 194, 16, 134, 88, 201, 251}
	digest, err := tx.Digest()
	require.NoError(t, err)
	require.Equal(t, targetHash, digest.Data())
}
