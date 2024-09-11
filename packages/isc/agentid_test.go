package isc

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/util/bcs"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

func TestAgentIDSerialization(t *testing.T) {
	n := &NilAgentID{}
	bcs.TestCodec(t, AgentID(n))
	rwutil.StringTest(t, AgentID(n), AgentIDFromString)

	a := NewAddressAgentID(cryptolib.NewRandomAddress())
	bcs.TestCodec(t, AgentID(a))
	rwutil.StringTest(t, AgentID(a), AgentIDFromString)
	rwutil.StringTest(t, a, addressAgentIDFromString)

	chainID := ChainIDFromAddress(cryptolib.NewRandomAddress())
	c := NewContractAgentID(chainID, 42)
	bcs.TestCodec(t, AgentID(c))
	rwutil.StringTest(t, AgentID(c), AgentIDFromString)

	e := NewEthereumAddressAgentID(chainID, common.HexToAddress("1074"))
	bcs.TestCodec(t, AgentID(e))
	rwutil.StringTest(t, AgentID(e), AgentIDFromString)
	rwutil.StringTest(t, AgentID(e), AgentIDFromString)
}
