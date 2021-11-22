package iscp

import (
	"bytes"
	"time"

	"github.com/iotaledger/wasp/packages/util"
	"golang.org/x/xerrors"

	"github.com/iotaledger/wasp/packages/hashing"
)

// StateData represent parsed data stored as a metadata in the anchor output
type StateData struct {
	Commitment hashing.HashValue
	Timestamp  time.Time
}

func StateDataFromBytes(data []byte) (StateData, error) {
	ret := StateData{}
	if len(data) != hashing.HashSize+8 {
		return ret, xerrors.New("StateDataFromBytes: wrong bytes")
	}
	ret.Commitment, _ = hashing.HashValueFromBytes(data[:hashing.HashSize])
	n, _ := util.Int64From8Bytes(data[hashing.HashSize:])
	ret.Timestamp = time.Unix(0, n)
	return ret, nil
}

func (s *StateData) Bytes() []byte {
	var buf bytes.Buffer

	buf.Write(s.Commitment[:])
	buf.Write(util.Int64To8Bytes(s.Timestamp.UnixNano()))
	return buf.Bytes()
}
