package crypto

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.dedis.ch/kyber/v3/share"
	"go.dedis.ch/kyber/v3/suites"
)

var (
	suite = suites.MustFind("ed25519")

	secret = suite.Scalar().SetInt64(42)
)

func TestDecryptShare(t *testing.T) {
	dealerPrivKey := suite.Scalar().Pick(suite.RandomStream())
	dealerPubKey := suite.Point().Mul(dealerPrivKey, G)

	peerPrivKey := suite.Scalar().Pick(suite.RandomStream())
	peerPubKey := suite.Point().Mul(peerPrivKey, G)

	// manually create a deal
	var deal Deal
	poly := share.NewPriPoly(suite, 1, secret, suite.RandomStream())
	_, deal.Commits = poly.Commit(nil).Info()
	deal.PubKey = dealerPubKey
	salt, _ := deal.Commits.MarshalBinary()
	aead := newAEAD(Secret(suite, peerPubKey, dealerPrivKey), salt, contextInfo(0))
	deal.Shares = [][]byte{encryptScalar(poly.Eval(0).V, aead)}
	require.Len(t, deal.Shares[0], ShareLen(suite))

	s, err := DecryptShare(suite, &deal, 0, Secret(suite, dealerPubKey, peerPrivKey))
	require.NoError(t, err)
	require.Equal(t, &share.PriShare{I: 0, V: secret}, s)

	// decryption fails
	deal.Shares[0][ShareLen(suite)-1]++
	_, err = DecryptShare(suite, &deal, 0, Secret(suite, dealerPubKey, peerPrivKey))
	require.ErrorIs(t, err, ErrDecryptionFailed)

	// verification fails
	deal.Shares[0] = encryptScalar(suite.Scalar().Zero(), aead)
	_, err = DecryptShare(suite, &deal, 0, Secret(suite, dealerPubKey, peerPrivKey))
	require.ErrorIs(t, err, ErrVerificationFailed)
}
