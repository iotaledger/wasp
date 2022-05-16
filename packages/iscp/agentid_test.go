package iscp

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/stretchr/testify/require"
)

var networkPrefix = iotago.PrefixDevnet

func TestAgentID(t *testing.T) {
	{
		n := &NilAgentID{}

		{
			require.Equal(t, "-", n.String(networkPrefix))
			n2, err := NewAgentIDFromString("-", networkPrefix)
			require.NoError(t, err)
			require.EqualValues(t, n, n2)
			require.True(t, n.Equals(n2))
		}

		{
			b := n.Bytes()
			require.Len(t, b, 1)
			n2, err := AgentIDFromBytes(b)
			require.NoError(t, err)
			require.EqualValues(t, n, n2)
			require.True(t, n.Equals(n2))
		}
	}

	{
		a := NewAgentID(tpkg.RandEd25519Address())

		{
			s := a.String(networkPrefix)
			require.NotEqual(t, "-", s)
			require.NotContains(t, s, "@")
			require.Equal(t, string(networkPrefix), s[:len(networkPrefix)])
			a2, err := NewAgentIDFromString(s, networkPrefix)
			require.NoError(t, err)
			require.EqualValues(t, a, a2)
			require.True(t, a.Equals(a2))
		}

		{
			b := a.Bytes()
			a2, err := AgentIDFromBytes(b)
			require.NoError(t, err)
			require.EqualValues(t, a, a2)
			require.True(t, a.Equals(a2))
		}
	}

	{
		chid := ChainIDFromAddress(tpkg.RandAliasAddress())
		a := NewContractAgentID(&chid, 42)

		{
			s := a.String(networkPrefix)
			require.Contains(t, s, "@")
			require.NotContains(t, s, string(networkPrefix))
			a2, err := NewAgentIDFromString(s, networkPrefix)
			require.NoError(t, err)
			require.EqualValues(t, a, a2)
			require.True(t, a.Equals(a2))
		}

		{
			b := a.Bytes()
			a2, err := AgentIDFromBytes(b)
			require.NoError(t, err)
			require.EqualValues(t, a, a2)
			require.True(t, a.Equals(a2))
		}
	}

	{
		a := NewEthereumAddressAgentID(common.HexToAddress("1074"))

		{
			s := a.String(networkPrefix)
			require.NotContains(t, s, "@")
			require.Regexp(t, `^0x[^@]+`, s)
			a2, err := NewAgentIDFromString(s, networkPrefix)
			require.NoError(t, err)
			require.EqualValues(t, a, a2)
			require.True(t, a.Equals(a2))
		}

		{
			b := a.Bytes()
			a2, err := AgentIDFromBytes(b)
			require.NoError(t, err)
			require.EqualValues(t, a, a2)
			require.True(t, a.Equals(a2))
		}
	}
}
