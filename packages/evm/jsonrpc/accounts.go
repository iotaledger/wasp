// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package jsonrpc

import (
	"crypto/ecdsa"
	"slices"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type AccountManager struct {
	accounts map[common.Address]*ecdsa.PrivateKey
	addrs    []common.Address
}

func NewAccountManager(accounts []*ecdsa.PrivateKey) *AccountManager {
	a := &AccountManager{
		accounts: make(map[common.Address]*ecdsa.PrivateKey),
	}
	for _, account := range accounts {
		a.Add(account)
	}
	return a
}

func (a *AccountManager) Add(keyPair *ecdsa.PrivateKey) {
	addr := crypto.PubkeyToAddress(keyPair.PublicKey)
	if _, ok := a.accounts[addr]; ok {
		return
	}
	a.accounts[addr] = keyPair
	a.addrs = append(a.addrs, addr)
}

func (a *AccountManager) Get(addr common.Address) *ecdsa.PrivateKey {
	return a.accounts[addr]
}

func (a *AccountManager) Addresses() []common.Address {
	return slices.Clone(a.addrs)
}
