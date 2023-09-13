package isc

import (
	"io"

	"github.com/iotaledger/wasp/packages/util/rwutil"
)

const nilAgentIDString = "-"

type NilAgentID struct{}

var _ AgentID = &NilAgentID{}

func (a *NilAgentID) Bytes() []byte {
	return rwutil.WriteToBytes(a)
}

func (a *NilAgentID) BelongsToChain(cID ChainID) bool {
	return false
}

func (a *NilAgentID) BytesWithoutChainID() []byte {
	return a.Bytes()
}

func (a *NilAgentID) Equals(other AgentID) bool {
	if other == nil {
		return false
	}
	return other.Kind() == a.Kind()
}

func (a *NilAgentID) Kind() AgentIDKind {
	return AgentIDKindNil
}

func (a *NilAgentID) String() string {
	return nilAgentIDString
}

func (a *NilAgentID) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	rr.ReadKindAndVerify(rwutil.Kind(a.Kind()))
	return rr.Err
}

func (a *NilAgentID) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteKind(rwutil.Kind(a.Kind()))
	return ww.Err
}
