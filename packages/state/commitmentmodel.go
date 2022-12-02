package state

import (
	"bytes"
	"errors"

	"github.com/iotaledger/trie.go/common"
	"github.com/iotaledger/trie.go/models/trie_blake2b"
	"github.com/iotaledger/trie.go/models/trie_blake2b/trie_blake2b_verify"
)

var (
	commitmentModel      = trie_blake2b.New(common.PathArity16, trie_blake2b.HashSize160, 64)
	vectorCommitmentSize = len(commitmentModel.NewVectorCommitment().Bytes())
)

func EqualCommitments(c1, c2 common.Serializable) bool {
	return commitmentModel.EqualCommitments(c1, c2)
}

func VCommitmentFromBytes(data []byte) (common.VCommitment, error) {
	if len(data) != vectorCommitmentSize {
		return nil, errors.New("wrong data size")
	}
	ret := commitmentModel.NewVectorCommitment()
	if err := ret.Read(bytes.NewReader(data)); err != nil {
		return nil, err
	}
	return ret, nil
}

func ValidateMerkleProof(proof *trie_blake2b.MerkleProof, root common.VCommitment, value ...[]byte) error {
	if len(value) == 0 {
		return trie_blake2b_verify.Validate(proof, root.Bytes())
	}
	tc := commitmentModel.CommitToData(value[0])
	return trie_blake2b_verify.ValidateWithTerminal(proof, root.Bytes(), tc.Bytes())
}
