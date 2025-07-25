package trie

import (
	"crypto/rand"
	"encoding/hex"
	"io"

	"github.com/iotaledger/wasp/v2/packages/util/rwutil"
)

const (
	HashSizeBits  = 160
	HashSizeBytes = HashSizeBits / 8

	vectorLength       = NumChildren + 2 // 16 children + terminal + path extension
	terminalIndex      = NumChildren
	pathExtensionIndex = NumChildren + 1
)

// compressToHashSize hashes data if longer than hash size, otherwise copies it
func compressToHashSize(data []byte) (ret []byte) {
	if len(data) <= HashSizeBytes {
		ret = make([]byte, len(data))
		copy(ret, data)
	} else {
		hash := blake2b160(data)
		ret = hash[:]
	}
	return
}

// hashVector is used to calculate the hash of a trie node
type hashVector [vectorLength][]byte

func (hashes *hashVector) Hash() Hash {
	sum := 0
	for _, b := range hashes {
		sum += len(b)
	}
	buf := make([]byte, 0, sum+len(hashes))
	for _, b := range hashes {
		buf = append(buf, byte(len(b)))
		buf = append(buf, b...)
	}
	return blake2b160(buf)
}

// Hash is a blake2b 160 bit (20 bytes) hash
type Hash [HashSizeBytes]byte

func HashFromBytes(data []byte) (ret Hash, err error) {
	_, err = rwutil.ReadFromBytes(data, &ret)
	return ret, err
}

func (h Hash) Bytes() []byte {
	return h[:]
}

func (h Hash) Clone() (ret Hash) {
	copy(ret[:], h[:])
	return
}

func (h Hash) Equals(other Hash) bool {
	return h == other
}

func (h Hash) String() string {
	return hex.EncodeToString(h[:])
}

func (h *Hash) Read(r io.Reader) error {
	return rwutil.ReadN(r, h[:])
}

func (h Hash) Write(w io.Writer) error {
	return rwutil.WriteN(w, h[:])
}

func RandomHash() Hash {
	var h Hash
	_, _ = rand.Read(h[:])
	return h
}
