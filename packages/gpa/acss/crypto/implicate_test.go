package crypto

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	G = suite.Point().Base()
	H = suite.Point().Mul(suite.Scalar().SetInt64(2), G)
)

func TestDLEQ(t *testing.T) {
	P1, P2 := suite.Point().Mul(secret, G), suite.Point().Mul(secret, H)

	t.Logf("proof that log_{G}(%s) == log_{H}(%s) == %s", P1, P2, secret)
	s, R1, R2 := dleqProof(suite, G, H, secret)
	require.True(t, dleqVerify(suite, G, H, P1, P2, s, R1, R2))

	// it must not validate for different points
	require.False(t, dleqVerify(suite, G, H, P1, H, s, R1, R2))
	require.False(t, dleqVerify(suite, G, H, G, P2, s, R1, R2))
}

func TestImplicate(t *testing.T) {
	private := suite.Scalar().SetInt64(42)
	public := suite.Point().Mul(private, G)

	data := Implicate(suite, H, private)
	require.Len(t, data, ImplicateLen(suite))

	secret, err := CheckImplicate(suite, H, public, data)
	require.NoError(t, err)
	require.Equal(t, Secret(suite, H, private), secret)
}
