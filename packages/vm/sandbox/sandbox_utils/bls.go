// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package sandbox_utils

import (
	"fmt"
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/crypto/bls"
	"go.dedis.ch/kyber/v3/pairing/bn256"
	"go.dedis.ch/kyber/v3/sign/bdn"
)

type blsUtil struct {
}

var suite = bn256.NewSuite()

func (u blsUtil) ValidSignature(data []byte, pubKeyBin []byte, signature []byte) bool {
	pubKey := suite.G2().Point()
	var err error
	if err = pubKey.UnmarshalBinary(pubKeyBin); err != nil {
		return false
	}
	return bdn.Verify(suite, pubKey, data, signature) == nil
}

func (u blsUtil) AddressFromPublicKey(pubKeyBin []byte) (ledgerstate.Address, error) {
	pubKey := suite.G2().Point()
	if err := pubKey.UnmarshalBinary(pubKeyBin); err != nil {
		return nil, fmt.Errorf("BLSUtil: wrong public key bytes")
	}
	return ledgerstate.NewBLSAddress(pubKeyBin), nil
}

// AggregateBLSSignatures
// TODO: optimize redundant binary manipulation.
//   Implement more flexible access to parts of SignatureShort
func (u blsUtil) AggregateBLSSignatures(pubKeysBin [][]byte, sigsBin [][]byte) ([]byte, []byte, error) {
	if len(sigsBin) == 0 || len(pubKeysBin) != len(sigsBin) {
		return nil, nil, fmt.Errorf("BLSUtil: number of public keys must be equal to the number of signatures and not empty")
	}

	sigPubKey := make([]bls.SignatureWithPublicKey, len(pubKeysBin))
	for i := range pubKeysBin {
		pubKey, _, err := bls.PublicKeyFromBytes(pubKeysBin[i])
		if err != nil {
			return nil, nil, fmt.Errorf("BLSUtil: wrong public key bytes: %v", err)
		}
		sig, _, err := bls.SignatureFromBytes(sigsBin[i])
		if err != nil {
			return nil, nil, fmt.Errorf("BLSUtil: wrong signature bytes: %v", err)
		}
		sigPubKey[i] = bls.NewSignatureWithPublicKey(pubKey, sig)
	}

	aggregatedSignature, err := bls.AggregateSignatures(sigPubKey...)
	if err != nil {
		return nil, nil, fmt.Errorf("BLSUtil: fialed aggregate signatures: %v", err)
	}

	return aggregatedSignature.PublicKey.Bytes(), aggregatedSignature.Signature.Bytes(), nil
}
