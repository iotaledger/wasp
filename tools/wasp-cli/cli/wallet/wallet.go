package wallet

import (
	"errors"
	"fmt"

	"github.com/iotaledger/wasp/clients/iota-go/hw_ledger"
	ledger_go "github.com/iotaledger/wasp/clients/iota-go/hw_ledger/ledger-go"
	"github.com/iotaledger/wasp/clients/iota-go/iotasigner"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/wallet/providers"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/wallet/wallets"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

var AddressIndex uint32

type WalletProvider string

const (
	ProviderReadOnlyMnemonic          WalletProvider = "unsafe_inmemory_testing_mnemonic"
	ProviderUnsafeInMemoryTestingSeed WalletProvider = "unsafe_inmemory_testing_seed"
	ProviderKeyChain                  WalletProvider = "keychain"
	ProviderLedger                    WalletProvider = "ledger"
	ProviderLedgerDebug               WalletProvider = "ledger_debug"
)

func GetWalletProvider() WalletProvider {
	provider := WalletProvider(config.GetWalletProviderString())

	switch provider {
	case ProviderKeyChain, ProviderUnsafeInMemoryTestingSeed, ProviderLedger, ProviderReadOnlyMnemonic, ProviderLedgerDebug:
		return provider
	}
	return ProviderKeyChain
}

func SetWalletProvider(provider WalletProvider) error {
	switch provider {
	case ProviderKeyChain, ProviderUnsafeInMemoryTestingSeed, ProviderLedger, ProviderLedgerDebug, ProviderReadOnlyMnemonic:
		config.SetWalletProviderString(string(provider))
		return nil
	}
	return errors.New("invalid wallet provider provided")
}

var loadedWallet wallets.Wallet

func initializeLedger(walletProvider WalletProvider) *hw_ledger.HWLedger {
	var ledgerDevice *hw_ledger.HWLedger
	var err error

	switch walletProvider {
	case ProviderLedgerDebug:
		log.Printf("Trying to open Speculos debug mode\n")

		dev, err2 := ledger_go.NewSpeculosTransport(ledger_go.DefaultSpeculosTransportOpts())
		log.Check(err2)
		ledgerDevice = hw_ledger.NewHWLedger(dev)
	case ProviderLedger:
		log.Printf("Trying to open Ledger\n")

		ledgerDevice, err = hw_ledger.TryAndConnect()
		log.Check(err)
	default:
		panic("incorrect wallet provider")
	}

	return ledgerDevice
}

func Load() wallets.Wallet {
	walletProvider := GetWalletProvider()

	if loadedWallet == nil {
		switch walletProvider {
		case ProviderLedger, ProviderLedgerDebug:
			loadedWallet = providers.NewExternalWallet(initializeLedger(walletProvider), AddressIndex, iotasigner.IotaCoinType)
		case ProviderKeyChain:
			loadedWallet = providers.LoadKeyChain(AddressIndex)
		case ProviderUnsafeInMemoryTestingSeed:
			loadedWallet = providers.LoadUnsafeInMemoryTestingSeed(AddressIndex)
		case ProviderReadOnlyMnemonic:
			loadedWallet = providers.LoadMnemonicSeed()

		}
	}

	return loadedWallet
}

func InitWallet(overwrite bool) {
	walletProvider := GetWalletProvider()

	switch walletProvider {
	case ProviderKeyChain:
		providers.CreateKeyChain(overwrite)
	case ProviderUnsafeInMemoryTestingSeed:
		providers.CreateUnsafeInMemoryTestingSeed()
	}
}

func Migrate(provider WalletProvider) {
	seedHex := config.GetSeedForMigration()
	if seedHex == "" {
		fmt.Println("No seed to migrate found.")
		return
	}

	seedBytes, err := cryptolib.DecodeHex(seedHex)
	log.Check(err)
	seed := cryptolib.SeedFromBytes(seedBytes)

	switch provider {
	case ProviderKeyChain:
		providers.MigrateKeyChain(seed)
	default:
		log.Printf("Migration unsupported for provider %v", provider)
	}
}
