package cryptolib

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/require"

	suisigner2 "github.com/iotaledger/wasp/clients/iota-go/suisigner"
)

func TestCompatability(t *testing.T) {
	seedIndex := uint32(1)
	seed := make([]byte, 32)
	n, err := rand.Read(seed)
	require.NoError(t, err)
	require.Equal(t, len(seed), n)

	kp := KeyPairFromSeed(SubSeed(seed, seedIndex))
	subseed := SubSeed(seed, seedIndex)
	suikp := suisigner2.NewSigner(subseed[:], suisigner2.KeySchemeFlagIotaEd25519)

	require.Equal(t, kp.Address().AsSuiAddress().Data(), suikp.Address().Data())

	kpSign, err := kp.SignTransactionBlock([]byte{1, 2, 3, 4}, suisigner2.DefaultIntent())
	require.NoError(t, err)

	suiSign, err := suikp.SignTransactionBlock([]byte{1, 2, 3, 4}, suisigner2.DefaultIntent())
	require.NoError(t, err)

	require.Equal(t, kpSign.AsSuiSignature().Bytes(), suiSign.Bytes())
}
