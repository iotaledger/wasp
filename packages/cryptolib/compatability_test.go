package cryptolib

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/suisigner"
)

const (
	SEEDFORUSER = "0x1234"
)

func TestCompatability(t *testing.T) {
	seedIndex := uint32(1)
	seed := SEEDFORUSER

	kp := KeyPairFromSeed(SubSeed(sui.MustAddressFromHex(seed)[:], seedIndex))
	subseed := SubSeed(sui.MustAddressFromHex(seed)[:], seedIndex)
	suikp := suisigner.NewSigner(subseed[:], suisigner.KeySchemeFlagEd25519)

	require.Equal(t, kp.Address().AsSuiAddress().Data(), suikp.Address().Data())

	kpSign, err := kp.SignTransactionBlock([]byte{1, 2, 3, 4}, suisigner.DefaultIntent())
	require.NoError(t, err)

	suiSign, err := suikp.SignTransactionBlock([]byte{1, 2, 3, 4}, suisigner.DefaultIntent())
	require.NoError(t, err)

	require.Equal(t, kpSign.AsSuiSignature().Bytes(), suiSign.Bytes())
}
