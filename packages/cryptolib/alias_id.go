package cryptolib

import "golang.org/x/crypto/blake2b"

const (
	// AliasIDLength is the byte length of an AliasID.
	AliasIDLength = blake2b.Size256
)

// AliasID is the identifier for an alias account.
// It is computed as the Blake2b-256 hash of the OutputID of the output which created the account.
type AliasID [AliasIDLength]byte

func (id AliasID) String() string {
	return EncodeHex(id[:])
}
