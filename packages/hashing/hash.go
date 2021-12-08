package hashing

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"math/rand"

	"github.com/mr-tron/base58"
	"golang.org/x/crypto/sha3"

	// github.com/mr-tron/base58

	"github.com/pkg/errors"
	"golang.org/x/crypto/blake2b"
)

const HashSize = 32

type HashValue [HashSize]byte

var NilHash = HashValue{}

func (h HashValue) Bytes() []byte {
	return h[:]
}

func (h HashValue) String() string {
	return h.Hex()
}

func (h HashValue) Base58() string {
	return base58.Encode(h[:])
}

func (h HashValue) Hex() string {
	return hex.EncodeToString(h[:])
}

func (h *HashValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(h.String())
}

func (h *HashValue) UnmarshalJSON(buf []byte) error {
	var s string
	err := json.Unmarshal(buf, &s)
	if err != nil {
		return err
	}
	ret, err := HashValueFromBase58(s)
	if err != nil {
		return err
	}
	copy(h[:], ret[:])
	return nil
}

func HashValueFromBytes(b []byte) (HashValue, error) {
	if len(b) != HashSize {
		return NilHash, errors.New("wrong HashValue bytes length")
	}
	var ret HashValue
	copy(ret[:], b)
	return ret, nil
}

func HashValueFromBase58(s string) (HashValue, error) {
	b, err := base58.Decode(s)
	if err != nil {
		return NilHash, err
	}
	return HashValueFromBytes(b)
}

// HashData Blake2b
func HashData(data ...[]byte) HashValue {
	return HashDataBlake2b(data...)
}

func HashDataBlake2b(data ...[]byte) (ret HashValue) {
	h := hashBlake2b()
	for _, d := range data {
		_, err := h.Write(d)
		if err != nil {
			panic(err)
		}
	}
	copy(ret[:], h.Sum(nil))
	return
}

func hashBlake2b() hash.Hash {
	h, err := blake2b.New256(nil)
	if err != nil {
		panic(err)
	}
	if h.Size() != HashSize {
		panic("blake2b: hash size != 32")
	}
	return h
}

func HashSha3(data ...[]byte) (ret HashValue) {
	h := hashSha3()
	for _, d := range data {
		_, err := h.Write(d)
		if err != nil {
			panic(err)
		}
	}
	copy(ret[:], h.Sum(nil))
	return
}

func hashSha3() hash.Hash {
	h := sha3.New256()
	if h.Size() != HashSize {
		panic("sha3: hash size != 32")
	}
	return h
}

func HashStrings(str ...string) HashValue {
	tarr := make([][]byte, len(str))
	for i, s := range str {
		tarr[i] = []byte(s)
	}
	return HashData(tarr...)
}

func RandomHash(rnd *rand.Rand) HashValue {
	s := ""
	if rnd == nil {
		s = fmt.Sprintf("%d", rand.Int())
	} else {
		s = fmt.Sprintf("%d", rnd.Int())
	}
	ret := HashStrings(s, s, s)
	return ret
}

func (h *HashValue) Write(w io.Writer) error {
	_, err := w.Write(h[:])
	return err
}

func (h *HashValue) Read(r io.Reader) error {
	n, err := r.Read(h[:])
	if err != nil {
		return err
	}
	if n != HashSize {
		return errors.New("not enough bytes for HashValue")
	}
	return nil
}
