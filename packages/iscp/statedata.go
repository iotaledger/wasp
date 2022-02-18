package iscp

import (
	"bytes"
	"encoding/hex"

	"github.com/iotaledger/wasp/packages/hashing"
	"golang.org/x/xerrors"
)

type StateCommitment hashing.HashValue

// StateData represents the parsed data stored as a metadata in the anchor output
type StateData struct {
	Commitment StateCommitment
}

func StateDataFromBytes(data []byte) (StateData, error) {
	ret := StateData{}
	if len(data) != hashing.HashSize {
		return ret, xerrors.New("StateDataFromBytes: wrong bytes")
	}
	t, _ := hashing.HashValueFromBytes(data[:hashing.HashSize])
	ret.Commitment = (StateCommitment)(t)
	return ret, nil
}

func (s *StateData) Bytes() []byte {
	var buf bytes.Buffer

	buf.Write(s.Commitment[:])
	return buf.Bytes()
}

func (c StateCommitment) String() string {
	return (hashing.HashValue)(c).String()
}

const OriginStateCommitmentHex = "96yCdioNdifMb8xTeHQVQ8BzDnXDbRBoYzTq7iVaymvV"

func OriginStateCommitment() (ret StateCommitment) {
	retBin, err := hex.DecodeString(OriginStateCommitmentHex)
	if err != nil {
		panic(err)
	}
	copy(ret[:], retBin)
	return
}
