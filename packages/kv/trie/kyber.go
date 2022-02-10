package trie

import (
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing/bn256"
	"io"
)

type scalar struct {
	kyber.Scalar
}

type point struct {
	kyber.Point
}

var suite = bn256.NewSuite()

var KyberFactory = &CommitmentFactory{
	NewVectorCommitment:   newVectorCommitment,
	NewTerminalCommitment: newTerminalCommitment,
}

func NewKyberNode() *Node {
	return &Node{}
}

func newTerminalCommitment() TerminalCommitment {
	return scalar{Scalar: suite.G1().Scalar()}
}

func newVectorCommitment() VectorCommitment {
	return point{Point: suite.G1().Point()}
}

func (s scalar) Read(r io.Reader) error {
	_, err := s.UnmarshalFrom(r)
	return err
}

func (s scalar) Write(w io.Writer) {
	_, _ = s.MarshalTo(w)
}

func (p point) Read(r io.Reader) error {
	_, err := p.UnmarshalFrom(r)
	return err
}

func (p point) Write(w io.Writer) {
	_, _ = p.MarshalTo(w)
}
