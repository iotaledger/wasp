// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package sandbox_utils

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
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

func (u blsUtil) AddressFromPublicKey(pubKeyBin []byte) (address.Address, error) {
	pubKey := suite.G2().Point()
	if err := pubKey.UnmarshalBinary(pubKeyBin); err != nil {
		return address.Address{}, fmt.Errorf("BLSUtil: wrong public key bytes")
	}
	return address.FromBLSPubKey(pubKeyBin), nil
}

// AggregateBLSSignatures
// TODO: optimize redundant binary manipulation.
//   Implement more flexible access to parts of SignatureShort
func (u blsUtil) AggregateBLSSignatures(pubKeysBin [][]byte, sigsBin [][]byte) ([]byte, []byte, error) {
	if len(sigsBin) == 0 || len(pubKeysBin) != len(sigsBin) {
		return nil, nil, fmt.Errorf("BLSUtil: number of public keys must be equal to the number of signatures and not empty")
	}
	sigs := make([]signaturescheme.Signature, len(sigsBin))
	for i := range sigs {
		sigs[i] = signaturescheme.NewBLSSignature(pubKeysBin[i], sigsBin[i])
	}
	ret, err := signaturescheme.AggregateBLSSignatures(sigs...)
	if err != nil {
		return nil, nil, fmt.Errorf("BLSUtil: %v", err)
	}
	pubKeyBin := ret.Bytes()[1 : 1+signaturescheme.BLSPublicKeySize]
	sigBin := ret.Bytes()[1+signaturescheme.BLSPublicKeySize:]

	return pubKeyBin, sigBin, nil
}
