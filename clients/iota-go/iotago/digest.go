package iotago

type (
	Digest                   = Base58
	ObjectDigest             = Digest
	TransactionDigest        = Digest
	TransactionEffectsDigest = Digest
	TransactionEventsDigest  = Digest
	CheckpointDigest         = Digest
	CertificateDigest        = Digest
	CheckpointContentsDigest = Digest
)

func NewDigest(str string) (*Digest, error) {
	return NewBase58(str)
}

func MustNewDigest(str string) *Digest {
	digest, err := NewBase58(str)
	if err != nil {
		panic(err)
	}
	return digest
}

func (d Digest) Bytes() []byte {
	return d.Data()
}

func DigestFromBytes(b []byte) *Digest {
	ret := Digest(b)
	return &ret
}
