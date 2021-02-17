// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package sandbox_utils

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/hive.go/crypto/ed25519"
)

type ed25519Util struct {
}

func (u ed25519Util) ValidSignature(data []byte, pubKey []byte, signature []byte) (bool, error) {
	pk, _, err := ed25519.PublicKeyFromBytes(pubKey)
	if err != nil {
		return false, fmt.Errorf("ED255519Util: wrong public key bytes")
	}
	sig, _, err := ed25519.SignatureFromBytes(signature)
	if err != nil {
		return false, fmt.Errorf("ED255519Util: wrong signature bytes")
	}
	return pk.VerifySignature(data, sig), nil
}

func (u ed25519Util) AddressFromPublicKey(pubKey []byte) (address.Address, error) {
	pk, _, err := ed25519.PublicKeyFromBytes(pubKey)
	if err != nil {
		return address.Address{}, fmt.Errorf("ED255519Util: wrong public key bytes")
	}
	return address.FromED25519PubKey(pk), nil
}
