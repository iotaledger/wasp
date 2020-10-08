package testutil

import (
	"github.com/iotaledger/goshimmer/client/wallet/packages/seed"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/mr-tron/base58"
)

type Wallet struct {
	seed  *seed.Seed
	index uint64
}

func NewWallet(b58walletSeed string) *Wallet {
	seedBytes, err := base58.Decode(b58walletSeed)
	if err != nil {
		panic(err)
	}
	return &Wallet{seed: seed.NewSeed(seedBytes), index: 0}
}

func (w *Wallet) Address() *address.Address {
	addr := w.seed.Address(w.index).Address
	return &addr
}

func (w *Wallet) SigScheme() signaturescheme.SignatureScheme {
	return signaturescheme.ED25519(*w.seed.KeyPair(w.index))
}

func (w *Wallet) WithIndex(index int) *Wallet {
	return &Wallet{seed: w.seed, index: uint64(index)}
}
