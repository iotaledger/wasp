package isc

import (
	"errors"
	"io"

	"github.com/iotaledger/wasp/packages/util/rwutil"
)

const nilAgentIDString = "-"

type NilAgentID struct{}

var _ AgentID = &NilAgentID{}

func (a *NilAgentID) Kind() rwutil.Kind {
	return AgentIDKindNil
}

func (a *NilAgentID) Bytes() []byte {
	return rwutil.WriterToBytes(a)
}

func (a *NilAgentID) String() string {
	return nilAgentIDString
}

func (a *NilAgentID) Equals(other AgentID) bool {
	if other == nil {
		return false
	}
	return other.Kind() == a.Kind()
}

func (a *NilAgentID) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	kind := rwutil.Kind(rr.ReadByte())
	if rr.Err == nil && kind != a.Kind() {
		return errors.New("invalid NilAgentID kind")
	}
	return rr.Err
}

func (a *NilAgentID) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteUint8(uint8(a.Kind()))
	return ww.Err
}
