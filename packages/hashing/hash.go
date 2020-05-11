package hashing

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/mr-tron/base58"
	"io"

	// github.com/mr-tron/base58
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/sha3"
	"hash"
	"math/rand"
)

const HashSize = sha256.Size

type HashValue [HashSize]byte
type HashableBytes []byte

var nilHash HashValue
var NilHash = &nilHash

func (h *HashValue) Bytes() []byte {
	return (*h)[:]
}

func (h *HashValue) String() string {
	return base58.Encode(h[:])
	//return hex.EncodeToString(h[:])
}

func (h *HashValue) Short() string {
	return base58.Encode((*h)[:6]) + ".."
	//return hex.EncodeToString((*h)[:6]) + ".."
}

func (h *HashValue) Shortest() string {
	//return base58.Encode((*h)[:4])
	return hex.EncodeToString((*h)[:4])
}

func (h *HashValue) Equal(h1 *HashValue) bool {
	if h == h1 {
		return true
	}
	return *h == *h1
}

func (h *HashValue) Clone() *HashValue {
	var ret HashValue
	copy(ret[:], h.Bytes())
	return &ret
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
	ret, err := HashValueFromString(s)
	if err != nil {
		return err
	}
	copy(h.Bytes(), ret.Bytes())
	return nil
}

func HashValueFromString(s string) (*HashValue, error) {
	b, err := base58.Decode(s)
	//b, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}
	if len(b) != HashSize {
		return nil, errors.New("wrong hex encoded string. Can't convert to HashValue")
	}
	var ret HashValue
	copy(ret.Bytes(), b)
	return &ret, nil
}

func HashData(data ...[]byte) *HashValue {
	return HashDataBlake2b(data...)
	//return HashDataSha3(data...)
}

func HashDataBlake2b(data ...[]byte) *HashValue {
	h, err := blake2b.New256(nil)
	if err != nil {
		panic(err)
	}
	return hashTheData(h, data)
}

func HashDataSha3(data ...[]byte) *HashValue {
	h := sha3.New256()
	return hashTheData(h, data)
}

func hashTheData(h hash.Hash, data [][]byte) *HashValue {
	for _, d := range data {
		h.Write(d)
	}
	var ret HashValue
	copy(ret[:], h.Sum(nil))
	return &ret
}

func HashStrings(str ...string) *HashValue {
	tarr := make([][]byte, len(str))
	for i, s := range str {
		tarr[i] = []byte(s)
	}
	return HashData(tarr...)
}

func RandomHash(rnd *rand.Rand) *HashValue {
	s := ""
	if rnd == nil {
		s = fmt.Sprintf("%d", rand.Int())
	} else {
		s = fmt.Sprintf("%d", rnd.Int())
	}
	return HashStrings(s, s, s)
}

func HashInList(h *HashValue, list []*HashValue) bool {
	for _, h1 := range list {
		if h.Equal(h1) {
			return true
		}
	}
	return false
}

func (h *HashValue) Write(w io.Writer) error {
	_, err := w.Write(h.Bytes())
	return err
}

func (h *HashValue) Read(r io.Reader) error {
	n, err := r.Read(h.Bytes())
	if err != nil {
		return err
	}
	if n != HashSize {
		return errors.New("not enough bytes for HashValue")
	}
	return nil
}
