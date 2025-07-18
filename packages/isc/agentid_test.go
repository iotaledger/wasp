package isc

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

func TestAgentIDSerialization(t *testing.T) {
	n := &NilAgentID{}
	bcs.TestCodecAndHash(t, AgentID(n), "ec7fce86b53c")
	rwutil.StringTest(t, AgentID(n), AgentIDFromString)

	a := NewAddressAgentID(cryptolib.NewRandomAddress())
	bcs.TestCodec(t, AgentID(a))
	rwutil.StringTest(t, AgentID(a), AgentIDFromString)
	rwutil.StringTest(t, a, addressAgentIDFromString)

	a = NewAddressAgentID(cryptolib.TestAddress)
	bcs.TestCodecAndHash(t, AgentID(a), "db416c67f079")

	ChainIDFromAddress(cryptolib.NewRandomAddress())
	c := NewContractAgentID(42)
	bcs.TestCodec(t, AgentID(c))
	rwutil.StringTest(t, AgentID(c), AgentIDFromString)

	c = NewContractAgentID(42)
	bcs.TestCodecAndHash(t, AgentID(c), "2b77cf327574")

	e := NewEthereumAddressAgentID(common.HexToAddress("1074"))
	bcs.TestCodecAndHash(t, AgentID(e), "d77145f60e5d")
	rwutil.StringTest(t, AgentID(e), AgentIDFromString)
}
