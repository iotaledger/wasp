package trie

import (
	"io"
)

// Commitment is the common interface for VCommitment and TCommitment
type Commitment interface {
	Read(r io.Reader) error
	Write(w io.Writer) error
	Bytes() []byte
	String() string
}

// VCommitment (vector commitment) is a 20 bytes hash commitment to a trie node
type VCommitment interface {
	Commitment
	Hash() Hash
	Equals(VCommitment) bool
	Clone() VCommitment
}

// TCommitment (terminal commitment) is the commitment to the data stored in a
// trie node.
type TCommitment interface {
	Commitment
	ExtractValue() ([]byte, bool)
	Equals(TCommitment) bool
	Clone() TCommitment
}
