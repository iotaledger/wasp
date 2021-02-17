package coretypes

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/packages/hashing"
)

// Utils implement various utilities which are faster on host side than on wasm VM
// Implement deterministic stateless computations
type Utils interface {
	Base58() Base58
	Hashing() Hashing
	ED25519() ED25519
	BLS() BLS
}

type Hashing interface {
	Blake2b(data []byte) hashing.HashValue
	Sha3(data []byte) hashing.HashValue
	Hname(name string) Hname
}

type Base58 interface {
	Decode(s string) ([]byte, error)
	Encode(data []byte) string
}

type ED25519 interface {
	ValidSignature(data []byte, pubKey []byte, signature []byte) (bool, error)
	AddressFromPublicKey(pubKey []byte) (address.Address, error)
}

type BLS interface {
	ValidSignature(data []byte, pubKey []byte, signature []byte) (bool, error)
	AddressFromPublicKey(pubKey []byte) (address.Address, error)
	AggregateBLSSignatures(pubKeysBin [][]byte, sigsBin [][]byte) ([]byte, []byte, error)
}
