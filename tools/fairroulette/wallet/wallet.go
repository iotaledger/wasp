package wallet

import (
	"fmt"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/wallet"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/mr-tron/base58"
	"github.com/spf13/viper"
)

type WalletConfig struct {
	Seed []byte
}

type Wallet struct {
	goshimmerWallet *wallet.Wallet
}

func Init() error {
	seed := wallet.New().Seed().Bytes()
	viper.SetDefault("wallet.seed", base58.Encode(seed))
	return viper.WriteConfig()
}

func Load() *Wallet {
	seedb58 := viper.GetString("wallet.seed")
	if len(seedb58) == 0 {
		check(fmt.Errorf("call `wallet init` first"))
	}
	seed, err := base58.Decode(seedb58)
	check(err)
	return &Wallet{wallet.New(seed)}
}

var addressIndex int

func (w *Wallet) KeyPair() *ed25519.KeyPair {
	return w.goshimmerWallet.Seed().KeyPair(uint64(addressIndex))
}

func (w *Wallet) Address() address.Address {
	return w.goshimmerWallet.Seed().Address(uint64(addressIndex))
}

func (w *Wallet) SignatureScheme() signaturescheme.SignatureScheme {
	return signaturescheme.ED25519(*w.KeyPair())
}
