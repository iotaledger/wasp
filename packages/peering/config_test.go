package peering

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfigBasic(t *testing.T) {
	peers := []string{"localhost:4000", "localhost:4001", "localhost:4002", "localhost:4003"}
	t.Run("happy path", func(t *testing.T) {
		c, err := NewStaticPeerNetworkConfigProvider(peers[0], 4000, peers[1:]...)
		require.NoError(t, err)
		require.EqualValues(t, peers[0], c.OwnNetID())
	})
	t.Run("eliminate own netid", func(t *testing.T) {
		c, err := NewStaticPeerNetworkConfigProvider(peers[1], 4001, peers...)
		require.NoError(t, err)
		require.EqualValues(t, len(peers)-1, len(c.Neighbors()))
	})
	t.Run("wrong port", func(t *testing.T) {
		_, err := NewStaticPeerNetworkConfigProvider(peers[0], 4001, peers...)
		require.Error(t, err)
	})
	t.Run("duplicates", func(t *testing.T) {
		_, err := NewStaticPeerNetworkConfigProvider(peers[0], 4000, peers[1], peers[2], peers[1])
		require.Error(t, err)
	})
}
