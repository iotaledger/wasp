package isc

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"

	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

func TestAgentIDSerialization(t *testing.T) {
	n := &NilAgentID{}
	rwutil.BytesTest(t, AgentID(n), AgentIDFromBytes)
	rwutil.StringTest(t, AgentID(n), AgentIDFromString)

	a := NewAddressAgentID(tpkg.RandEd25519Address())
	rwutil.BytesTest(t, AgentID(a), AgentIDFromBytes)
	rwutil.StringTest(t, AgentID(a), AgentIDFromString)
	rwutil.StringTest(t, a, addressAgentIDFromString)

	chainID := ChainIDFromAddress(tpkg.RandAliasAddress())
	c := NewContractAgentID(chainID, 42)
	rwutil.BytesTest(t, AgentID(c), AgentIDFromBytes)
	rwutil.StringTest(t, AgentID(c), AgentIDFromString)

	e := NewEthereumAddressAgentID(chainID, common.HexToAddress("1074"))
	rwutil.BytesTest(t, AgentID(e), AgentIDFromBytes)
	rwutil.StringTest(t, AgentID(e), AgentIDFromString)
	rwutil.StringTest(t, AgentID(e), AgentIDFromString)
}
