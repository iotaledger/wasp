package cryptolib

import (
	crypto "crypto/ed25519"

	"github.com/iotaledger/hive.go/crypto/ed25519"
	iotago "github.com/iotaledger/iota.go/v3"
)

func Ed25519AddressFromPubKey(key PublicKey) *iotago.Ed25519Address {
	ret := iotago.Ed25519AddressFromPubKey(key)
	return &ret
}

func Verify(publicKey PublicKey, message, sig []byte) bool {
	return crypto.Verify(publicKey, message, sig)
}

func CryptolibPublicKeyToHivePublicKey(pk PublicKey) ed25519.PublicKey { // nolint:revive
	var result ed25519.PublicKey
	copy(result[:], pk)
	return result
}

func CryptolibPrivateKeyToHivePrivateKey(pk PrivateKey) ed25519.PrivateKey { // nolint:revive
	var result ed25519.PrivateKey
	copy(result[:], pk)
	return result
}

func CryptolibKeyPairToHiveKeyPair(pk KeyPair) ed25519.KeyPair { // nolint:revive
	return ed25519.KeyPair{
		PrivateKey: CryptolibPrivateKeyToHivePrivateKey(pk.PrivateKey),
		PublicKey:  CryptolibPublicKeyToHivePublicKey(pk.PublicKey),
	}
}

func HivePublicKeyToCryptolibPublicKey(pk ed25519.PublicKey) PublicKey {
	result := make([]byte, len(pk))
	copy(result, pk[:])
	return result
}

func HivePrivateKeyToCryptolibPrivateKey(pk ed25519.PrivateKey) PrivateKey {
	result := make([]byte, len(pk))
	copy(result, pk[:])
	return result
}

func HiveKeyPairToCryptolibKeyPair(pk ed25519.KeyPair) KeyPair {
	return KeyPair{
		PrivateKey: HivePrivateKeyToCryptolibPrivateKey(pk.PrivateKey),
		PublicKey:  HivePublicKeyToCryptolibPublicKey(pk.PublicKey),
	}
}

/*func Sign(privateKey PrivateKey, message []byte) []byte {
	return crypto.Sign(privateKey.asCrypto(), message)
}

func NewAddressKeysForEd25519Address(addr *iotago.Ed25519Address, prvKey PrivateKey) iotago.AddressKeys {
	return iotago.NewAddressKeysForEd25519Address(addr, prvKey.asCrypto())
}*/
