package trie

import (
	"bytes"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNode(t *testing.T) {
	n := NewKyberNode()
	var buf bytes.Buffer
	n.Write(&buf)
	t.Logf("size() = %d, size(serialize) = %d", Size(n), len(buf.Bytes()))
	require.EqualValues(t, Size(n), len(buf.Bytes()))

	nBack, err := KyberFactory.NodeFromBytes(buf.Bytes())
	require.NoError(t, err)
	require.EqualValues(t, buf.Bytes(), Bytes(nBack))
}
