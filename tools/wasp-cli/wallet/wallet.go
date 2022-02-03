package wallet

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/tools/wasp-cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/mr-tron/base58"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type WalletConfig struct {
	Seed []byte
}

type Wallet struct {
	KeyPair cryptolib.KeyPair
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new wallet",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		seed := cryptolib.NewSeed()
		seedString := base58.Encode(seed[:])
		viper.Set("wallet.seed", seedString)
		log.Check(viper.WriteConfig())

		log.Printf("Initialized wallet seed in %s\n", config.ConfigPath)
		log.Printf("\nIMPORTANT: wasp-cli is alpha phase. The seed is currently being stored " +
			"in a plain text file which is NOT secure. Do not use this seed to store funds " +
			"in the mainnet!\n")
		log.Verbosef("\nSeed: %s\n", seedString)
	},
}

func Load() *Wallet {
	seedb58 := viper.GetString("wallet.seed")
	if seedb58 == "" {
		log.Fatalf("call `init` first")
	}
	seedBytes, err := base58.Decode(seedb58)
	log.Check(err)
	seed := cryptolib.SeedFromByteArray(seedBytes)
	return &Wallet{cryptolib.NewKeyPairFromSeed(seed)}
}

var addressIndex int

func (w *Wallet) PrivateKey() *cryptolib.PrivateKey {
	return &w.KeyPair.PrivateKey
}

func (w *Wallet) Address() iotago.Address {
	return cryptolib.Ed25519AddressFromPubKey(w.KeyPair.PublicKey)
}
