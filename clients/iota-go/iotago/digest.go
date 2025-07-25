package iotago

import (
	"bytes"
	"encoding/json"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/packages/hashing"
)

const DigestSize = 32

type (
	Digest                   [DigestSize]byte
	ObjectDigest             = Digest
	TransactionDigest        = Digest
	TransactionEffectsDigest = Digest
	TransactionEventsDigest  = Digest
	CheckpointDigest         = Digest
	CertificateDigest        = Digest
	CheckpointContentsDigest = Digest
)

func NewDigest(str string) (*Digest, error) {
	var ret Digest
	b58, err := NewBase58(str)
	if err != nil {
		return nil, err
	}
	copy(ret[:], *b58)
	return &ret, nil
}

func MustNewDigest(str string) *Digest {
	digest, err := NewDigest(str)
	if err != nil {
		panic(err)
	}
	return digest
}

func (d Digest) HashValue() hashing.HashValue {
	return hashing.HashValue(d)
}

func (d Digest) Bytes() []byte {
	return d[:]
}

func DigestFromBytes(b []byte) *Digest {
	if len(b) != DigestSize {
		panic("invalid bytes for Digest")
	}
	ret := Digest(b)
	return &ret
}

func (d Digest) String() string {
	b58 := Base58(d.Bytes())
	return b58.String()
}

func (d Digest) Equals(other Digest) bool {
	return bytes.Equal(d[:], other[:])
}

func (d Digest) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *Digest) UnmarshalJSON(data []byte) error {
	str := ""
	err := json.Unmarshal(data, &str)
	if err != nil {
		return err
	}
	tmp, err := NewBase58(str)
	if err == nil {
		*d = *DigestFromBytes(tmp.Data())
	}
	return err
}

func (d Digest) MarshalBCS(e *bcs.Encoder) error {
	e.Encode(d[:])
	return nil
}

func (d *Digest) UnmarshalBCS(de *bcs.Decoder) error {
	b := bcs.Decode[[]byte](de)
	copy(d[:], b)
	return nil
}
