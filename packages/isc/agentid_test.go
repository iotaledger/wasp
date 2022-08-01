package isc

import (
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/stretchr/testify/require"
)

func TestAgentID(t *testing.T) {
	networkPrefix := parameters.L1.Protocol.Bech32HRP

	{
		n := &NilAgentID{}

		{
			require.Equal(t, "-", n.String())
			n2, err := NewAgentIDFromString("-")
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
			s := a.String()
			require.NotEqual(t, "-", s)
			require.NotContains(t, s, "@")
			require.Equal(t, string(networkPrefix), s[:len(networkPrefix)])
			a2, err := NewAgentIDFromString(s)
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
			s := a.String()
			require.Contains(t, s, "@")
			parts := strings.Split(s, "@")
			require.Len(t, parts, 2)
			require.Equal(t, string(networkPrefix), parts[1][:len(networkPrefix)])
			a2, err := NewAgentIDFromString(s)
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
			s := a.String()
			require.NotContains(t, s, "@")
			require.Regexp(t, `^0x[^@]+`, s)
			a2, err := NewAgentIDFromString(s)
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
