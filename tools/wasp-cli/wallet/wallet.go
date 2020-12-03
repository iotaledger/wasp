package wallet

import (
	"github.com/iotaledger/goshimmer/client/wallet/packages/seed"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/tools/wasp-cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/mr-tron/base58"
	"github.com/spf13/viper"
)

type WalletConfig struct {
	Seed []byte
}

type Wallet struct {
	seed *seed.Seed
}

func initCmd(args []string) {
	seed := base58.Encode(seed.NewSeed().Bytes())
	viper.Set("wallet.seed", seed)
	log.Check(viper.WriteConfig())

	log.Printf("Initialized wallet seed in %s\n", config.ConfigPath)
	log.Verbose("Seed: %s\n", seed)
}

func Load() *Wallet {
	seedb58 := viper.GetString("wallet.seed")
	if len(seedb58) == 0 {
		log.Fatal("call `init` first")
	}
	seedBytes, err := base58.Decode(seedb58)
	log.Check(err)
	return &Wallet{seed.NewSeed(seedBytes)}
}

var addressIndex int

func (w *Wallet) KeyPair() *ed25519.KeyPair {
	return w.seed.KeyPair(uint64(addressIndex))
}

func (w *Wallet) Address() address.Address {
	return w.seed.Address(uint64(addressIndex)).Address
}

func (w *Wallet) SignatureScheme() signaturescheme.SignatureScheme {
	return signaturescheme.ED25519(*w.KeyPair())
}
