package jsonrpc

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type AccountManager struct {
	accounts map[common.Address]*ecdsa.PrivateKey
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
}

func (a *AccountManager) Addresses() []common.Address {
	ret := make([]common.Address, len(a.accounts))
	i := 0
	for addr, _ := range a.accounts {
		ret[i] = addr
		i++
	}
	return ret
}
