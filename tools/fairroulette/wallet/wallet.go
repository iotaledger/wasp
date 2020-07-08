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
	*wallet.Wallet
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

func (w *Wallet) AccountIndex() int {
	return viper.GetInt("address-index")
}

func (w *Wallet) KeyPair() *ed25519.KeyPair {
	return w.Seed().KeyPair(uint64(w.AccountIndex()))
}

func (w *Wallet) Address() address.Address {
	return w.Seed().Address(uint64(w.AccountIndex()))
}

func (w *Wallet) SignatureScheme() signaturescheme.SignatureScheme {
	return signaturescheme.ED25519(*w.KeyPair())
}
