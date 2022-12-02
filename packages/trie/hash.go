package trie

import (
	"bytes"
	"encoding/hex"
	"io"

	"github.com/iotaledger/wasp/packages/util"
)

const (
	HashSizeBits  = 160
	HashSizeBytes = HashSizeBits / 8

	vectorLength       = NumChildren + 2 // 16 children + terminal + path extension
	terminalIndex      = NumChildren
	pathExtensionIndex = NumChildren + 1
)

// Hash is a blake2b 160 bit (20 bytes) hash
type Hash [HashSizeBytes]byte

// hashVector is used to calculate the hash of a trie node
type hashVector [vectorLength][]byte

// compressToHashSize hashes data if longer than hash size, otherwise copies it
func compressToHashSize(data []byte) (ret []byte, valueInCommitment bool) {
	if len(data) <= HashSizeBytes {
		ret = make([]byte, len(data))
		valueInCommitment = true
		copy(ret, data)
	} else {
		hash := blake2b160(data)
		ret = hash[:]
	}
	return
}

func (hashes *hashVector) Hash() Hash {
	buf := make([]byte, vectorLength*HashSizeBytes)
	for i, h := range hashes {
		if h == nil {
			continue
		}
		pos := i * HashSizeBytes
		copy(buf[pos:pos+HashSizeBytes], h)
	}
	return blake2b160(buf)
}

func (h Hash) Clone() (ret Hash) {
	copy(ret[:], h[:])
	return
}

func (h Hash) Bytes() []byte {
	return h[:]
}

func (h *Hash) Read(r io.Reader) error {
	_, err := r.Read(h[:])
	return err
}

func (h Hash) Write(w io.Writer) error {
	_, err := w.Write(h[:])
	return err
}

func (h Hash) String() string {
	return hex.EncodeToString(h[:])
}

func (h Hash) Equals(other util.Equatable) bool {
	h2, ok := other.(Hash)
	if !ok {
		return false
	}
	return h == h2
}

func ReadHash(r io.Reader) (ret Hash, err error) {
	err = ret.Read(r)
	return
}

func HashFromBytes(data []byte) (ret Hash, err error) {
	rdr := bytes.NewReader(data)
	ret, err = ReadHash(rdr)
	if err != nil {
		return
	}
	if rdr.Len() > 0 {
		return Hash{}, ErrNotAllBytesConsumed
	}
	return
}
