// This file contains functions for verification of the proofs of inclusion or absence
// in the trie with trie_blake2b commitment model. The package only depends on the commitment model
// implementation and the proof format it defines. The verification package is completely independent on
// the implementation of the Merkle tree (the trie)
//
// DISCLAIMER: THE FOLLOWING CODE IS SECURITY CRITICAL.
// ANY POTENTIAL BUG WHICH MAY LEAD TO FALSE POSITIVES OF PROOF VALIDITY CHECKS POTENTIALLY
// CREATES AN ATTACK VECTOR.
// THEREFORE, IT IS HIGHLY RECOMMENDED THE VERIFICATION CODE TO BE WRITTEN BY THE VERIFYING PARTY ITSELF,
// INSTEAD OF CLONING THIS PACKAGE. DO NOT TRUST ANYBODY BUT YOURSELF. IN ANY CASE, PERFORM A DETAILED
// AUDIT OF THE PROOF-VERIFYING CODE BEFORE USING IT
package trie

import (
	"bytes"
	"errors"
	"fmt"
)

// MustKeyWithTerminal returns key and terminal commitment the proof is about. It returns:
// - key
// - terminal commitment. If it is nil, the proof is a proof of absence
// It does not verify the proof, so this function should be used only after Validate()
func (p *MerkleProof) MustKeyWithTerminal() ([]byte, []byte) {
	if len(p.Path) == 0 {
		return nil, nil
	}
	lastElem := p.Path[len(p.Path)-1]
	switch {
	case isValidChildIndex(lastElem.ChildIndex):
		if lastElem.Children[lastElem.ChildIndex] != nil {
			panic("nil child commitment expected for proof of absence")
		}
		return p.Key, nil
	case lastElem.ChildIndex == terminalCommitmentIndex:
		if lastElem.Terminal == nil {
			return p.Key, nil
		}
		return p.Key, lastElem.Terminal
	case lastElem.ChildIndex == pathExtensionCommitmentIndex:
		return p.Key, nil
	}
	panic("wrong lastElem.ChildIndex")
}

// IsProofOfAbsence checks if it is proof of absence. MerkleProof that the trie commits to something else in the place
// where it would commit to the key if it would be present
func (p *MerkleProof) IsProofOfAbsence() bool {
	_, r := p.MustKeyWithTerminal()
	return len(r) == 0
}

// Validate check the proof against the provided root commitments
func (p *MerkleProof) Validate(rootBytes []byte) error {
	if len(p.Path) == 0 {
		if len(rootBytes) != 0 {
			return errors.New("proof is empty")
		}
		return nil
	}
	c, err := p.verify(0, 0)
	if err != nil {
		return err
	}
	if !bytes.Equal(c[:], rootBytes) {
		return errors.New("invalid proof: commitment not equal to the root")
	}
	return nil
}

// ValidateWithTerminal checks the proof and checks if the proof commits to the specific value
// The check is dependent on the commitment model because of valueOptimisationThreshold
func (p *MerkleProof) ValidateWithTerminal(rootBytes, terminalBytes []byte) error {
	if err := p.Validate(rootBytes); err != nil {
		return err
	}
	_, terminalBytesInProof := p.MustKeyWithTerminal()
	compressedTerm, _ := compressToHashSize(terminalBytes)
	if !bytes.Equal(compressedTerm, terminalBytesInProof) {
		return errors.New("key does not correspond to the given value commitment")
	}
	return nil
}

func (p *MerkleProof) verify(pathIdx, keyIdx int) (Hash, error) {
	assert(pathIdx < len(p.Path), "assertion: pathIdx < lenPlus1(p.Path)")
	assert(keyIdx <= len(p.Key), "assertion: keyIdx <= lenPlus1(p.Key)")

	elem := p.Path[pathIdx]
	tail := p.Key[keyIdx:]
	isPrefix := bytes.HasPrefix(tail, elem.PathExtension)
	last := pathIdx == len(p.Path)-1
	if !last && !isPrefix {
		return Hash{}, fmt.Errorf("wrong proof: proof path does not follow the key. Path position: %d, key position %d", pathIdx, keyIdx)
	}
	if !last {
		assert(isPrefix, "assertion: isPrefix")
		if !isValidChildIndex(elem.ChildIndex) {
			return Hash{}, fmt.Errorf("wrong proof: wrong child index. Path position: %d, key position %d", pathIdx, keyIdx)
		}
		if elem.Children[byte(elem.ChildIndex)] != nil {
			return Hash{}, fmt.Errorf("wrong proof: unexpected commitment at child index %d. Path position: %d, key position %d", elem.ChildIndex, pathIdx, keyIdx)
		}
		nextKeyIdx := keyIdx + len(elem.PathExtension) + 1
		if nextKeyIdx > len(p.Key) {
			return Hash{}, fmt.Errorf("wrong proof: proof path out of key bounds. Path position: %d, key position %d", pathIdx, keyIdx)
		}
		c, err := p.verify(pathIdx+1, nextKeyIdx)
		if err != nil {
			return Hash{}, err
		}
		return elem.hash(c[:])
	}
	// it is the last in the path
	if isValidChildIndex(elem.ChildIndex) {
		if elem.Children[byte(elem.ChildIndex)] != nil {
			return Hash{}, fmt.Errorf("wrong proof: child commitment of the last element expected to be nil. Path position: %d, key position %d", pathIdx, keyIdx)
		}
		return elem.hash(nil)
	}
	if elem.ChildIndex != terminalCommitmentIndex && elem.ChildIndex != pathExtensionCommitmentIndex {
		return Hash{}, fmt.Errorf("wrong proof: child index expected to be %d or %d. Path position: %d, key position %d",
			terminalCommitmentIndex, pathExtensionCommitmentIndex, pathIdx, keyIdx)
	}
	return elem.hash(nil)
}

const errTooLongCommitment = "too long commitment at position %d. Can't be longer than %d bytes"

func (e *MerkleProofElement) makeHashVector(missingCommitment []byte) (*hashVector, error) {
	hashes := &hashVector{}
	for idx, c := range e.Children {
		if c == nil {
			continue
		}
		if !isValidChildIndex(int(idx)) {
			return nil, fmt.Errorf("wrong child index %d", idx)
		}
		if len(c) > HashSizeBytes {
			return nil, fmt.Errorf(errTooLongCommitment, idx, HashSizeBytes)
		}
		hashes[idx] = c[:]
	}
	if len(e.Terminal) > 0 {
		if len(e.Terminal) > HashSizeBytes {
			return nil, fmt.Errorf(errTooLongCommitment+" (terminal)", terminalCommitmentIndex, HashSizeBytes)
		}
		hashes[terminalCommitmentIndex] = e.Terminal
	}
	rawBytes, _ := compressToHashSize(e.PathExtension)
	hashes[pathExtensionCommitmentIndex] = rawBytes
	if isValidChildIndex(e.ChildIndex) {
		if len(missingCommitment) > HashSizeBytes {
			return nil, fmt.Errorf(errTooLongCommitment+" (skipped commitment)", e.ChildIndex, HashSizeBytes)
		}
		hashes[e.ChildIndex] = missingCommitment
	}
	return hashes, nil
}

func (e *MerkleProofElement) hash(missingCommitment []byte) (Hash, error) {
	hashVector, err := e.makeHashVector(missingCommitment)
	if err != nil {
		return Hash{}, err
	}
	return hashVector.Hash(), nil
}

func (p *MerkleProof) ValidateValue(trieRoot VCommitment, value []byte) error {
	tc := CommitToData(value)
	return p.ValidateWithTerminal(trieRoot.Bytes(), tc.Bytes())
}
