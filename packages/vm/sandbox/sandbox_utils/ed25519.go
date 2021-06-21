// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package sandbox_utils

import (
	"fmt"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/crypto/ed25519"
)

type ed25519Util struct{}

func (u ed25519Util) ValidSignature(data []byte, pubKey []byte, signature []byte) bool {
	pk, _, err := ed25519.PublicKeyFromBytes(pubKey)
	if err != nil {
		return false
	}
	sig, _, err := ed25519.SignatureFromBytes(signature)
	if err != nil {
		return false
	}
	return pk.VerifySignature(data, sig)
}

func (u ed25519Util) AddressFromPublicKey(pubKey []byte) (ledgerstate.Address, error) {
	pk, _, err := ed25519.PublicKeyFromBytes(pubKey)
	if err != nil {
		return nil, fmt.Errorf("ED255519Util: wrong public key bytes. Err: %v", err)
	}
	return ledgerstate.NewED25519Address(pk), nil
}
