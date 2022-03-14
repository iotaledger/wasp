package tcrypto

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/util/random"

	iotago "github.com/iotaledger/iota.go/v3"

	"go.dedis.ch/kyber/v3/pairing/bn256"
)

func TestMarshaling(t *testing.T) {
	suite := bn256.NewSuite()
	randomness := random.New()

	rnd1 := make([]kyber.Point, 10)
	rnd2 := make([]kyber.Point, 10)

	for i := range rnd1 {
		rnd1[i] = suite.G2().Point().Pick(randomness)
		rnd2[i] = suite.G2().Point().Pick(randomness)
	}

	index := uint16(5)
	dks := &DKShareImpl{
		Address:       &iotago.AliasAddress{},
		Index:         &index,
		N:             10,
		T:             7,
		SharedPublic:  suite.G2().Point().Pick(randomness),
		PublicCommits: rnd1,
		PublicShares:  rnd2,
		PrivateShare:  suite.G2().Scalar().Pick(randomness),
		suite:         suite,
	}

	dksBack, err := DKShareFromBytes(dks.Bytes(), suite)
	require.NoError(t, err)
	require.EqualValues(t, dks.Bytes(), dksBack.Bytes())
}
