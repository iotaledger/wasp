// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmtest

import (
	"crypto/ecdsa"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

var (
	FaucetKey, _  = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
	FaucetAddress = crypto.PubkeyToAddress(FaucetKey.PublicKey)
	FaucetSupply  = new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(9))
)

// 10 random keys
var Accounts []*ecdsa.PrivateKey

func init() {
	for i := 0; i < 10; i++ {
		privateKey, err := crypto.GenerateKey()
		if err != nil {
			panic(err)
		}
		Accounts = append(Accounts, privateKey)
	}
}

func AccountAddress(i int) common.Address {
	return crypto.PubkeyToAddress(Accounts[i].PublicKey)
}
