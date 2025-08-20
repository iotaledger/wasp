package cryptolib

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotasigner"
)

func TestCompatability(t *testing.T) {
	seedIndex := uint32(1)
	seed := make([]byte, 32)
	n, err := rand.Read(seed)
	require.NoError(t, err)
	require.Equal(t, len(seed), n)

	kp := KeyPairFromSeed(SubSeed(seed, seedIndex))
	subseed := SubSeed(seed, seedIndex)
	suikp := iotasigner.NewSigner(subseed[:], iotasigner.KeySchemeFlagDefault)

	require.Equal(t, kp.Address().AsIotaAddress().Data(), suikp.Address().Data())

	kpSign, err := kp.SignTransactionBlock([]byte{1, 2, 3, 4}, iotasigner.DefaultIntent())
	require.NoError(t, err)

	suiSign, err := suikp.SignTransactionBlock([]byte{1, 2, 3, 4}, iotasigner.DefaultIntent())
	require.NoError(t, err)

	require.Equal(t, kpSign.AsIotaSignature().Bytes(), suiSign.Bytes())
}
