package crypto

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.dedis.ch/kyber/v3"
)

func TestNewDeal(t *testing.T) {
	private := suite.Scalar().Pick(suite.RandomStream())
	public := suite.Point().Mul(private, G)

	deal := NewDeal(suite, []kyber.Point{public}, secret)
	require.NotNil(t, deal)

	data, err := deal.MarshalBinary()
	require.NoError(t, err)
	require.Len(t, data, DealLen(suite, 1))

	deal2, err := DealUnmarshalBinary(suite, 1, data)
	require.NoError(t, err)
	require.True(t, deal.Commits[0].Equal(deal2.Commits[0]))
	require.True(t, deal.PubKey.Equal(deal2.PubKey))
	require.Equal(t, deal.Shares, deal2.Shares)
}
