package cryptolib

import (
	"fmt"

	"github.com/iotaledger/hive.go/crypto/ed25519"
)

const (
	SignatureSize = ed25519.SignatureSize
)

func SignatureFromBytes(bytes []byte) (result [SignatureSize]byte, consumedBytes int, err error) {
	if len(bytes) < SignatureSize {
		err = fmt.Errorf("bytes too short")
		return
	}

	copy(result[:SignatureSize], bytes)
	consumedBytes = SignatureSize

	return
}

// TODO: remove all these functions
/*func CryptolibPublicKeyToHivePublicKey(pk PublicKey) ed25519.PublicKey { // nolint:revive
	var result ed25519.PublicKey
	copy(result[:], pk.asCrypto())
	return result
}

func CryptolibPrivateKeyToHivePrivateKey(pk PrivateKey) ed25519.PrivateKey { // nolint:revive
	var result ed25519.PrivateKey
	copy(result[:], pk.asCrypto())
	return result
}

func CryptolibKeyPairToHiveKeyPair(pk KeyPair) ed25519.KeyPair { // nolint:revive
	return ed25519.KeyPair{
		PrivateKey: CryptolibPrivateKeyToHivePrivateKey(pk.privateKey),
		PublicKey:  CryptolibPublicKeyToHivePublicKey(pk.publicKey),
	}
}

func HivePublicKeyToCryptolibPublicKey(pk ed25519.PublicKey) PublicKey {
	result := make([]byte, len(pk))
	copy(result, pk[:])
	return PublicKey{result}
}

func HivePrivateKeyToCryptolibPrivateKey(pk ed25519.PrivateKey) PrivateKey {
	result := make([]byte, len(pk))
	copy(result, pk[:])
	return PrivateKey{result}
}

func HiveKeyPairToCryptolibKeyPair(pk ed25519.KeyPair) KeyPair {
	return KeyPair{
		privateKey: HivePrivateKeyToCryptolibPrivateKey(pk.PrivateKey),
		publicKey:  HivePublicKeyToCryptolibPublicKey(pk.PublicKey),
	}
}*/
