package wallet

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"

	iotago "github.com/iotaledger/iota.go/v3"
	wasp_wallet_sdk "github.com/iotaledger/wasp-wallet-sdk"
	"github.com/iotaledger/wasp-wallet-sdk/types"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/wallet/providers"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/wallet/wallets"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

var AddressIndex uint32

type WalletProvider string

const (
	ProviderKeyChain   WalletProvider = "keychain"
	ProviderLedger     WalletProvider = "sdk_ledger"
	ProviderStronghold WalletProvider = "sdk_stronghold"
)

func GetWalletProvider() WalletProvider {
	provider := WalletProvider(config.GetWalletProviderString())

	switch provider {
	case ProviderLedger, ProviderKeyChain, ProviderStronghold:
		return provider
	}
	return ProviderKeyChain
}

func SetWalletProvider(provider WalletProvider) error {
	switch provider {
	case ProviderLedger, ProviderKeyChain, ProviderStronghold:
		config.SetWalletProviderString(string(provider))
		return nil
	}
	return errors.New("invalid wallet provider provided")
}

func getIotaSDKLibName() string {
	switch runtime.GOOS {
	case "windows":
		return "iota_sdk.dll"
	case "linux":
		return "libiota_sdk.so"
	case "darwin":
		return "libiota_sdk.dylib"
	default:
		panic(fmt.Sprintf("unsupported OS: %s", runtime.GOOS))
	}
}

func initIotaSDK(libPath string) *wasp_wallet_sdk.IOTASDK {
	sdk, err := wasp_wallet_sdk.NewIotaSDK(libPath)
	log.Check(err)

	_, err = sdk.InitLogger(types.ILoggerConfig{
		LevelFilter: config.GetWalletLogLevel(),
	})
	log.Check(err)

	return sdk
}

func getIotaSDK() *wasp_wallet_sdk.IOTASDK {
	// LoadLibrary (windows) and dlLoad (linux) have different search path behaviors
	// For now, use a relative path - as it will eventually be shipped with a release.
	// TODO: Revisit once proper release structure is set up.

	ex, err := os.Executable()
	log.Check(err)
	executableDir := filepath.Dir(ex)

	wd, err := os.Getwd()
	log.Check(err)

	searchPaths := []string{
		path.Join(executableDir, getIotaSDKLibName()),
		path.Join(wd, getIotaSDKLibName()),
	}

	for _, searchPath := range searchPaths {
		if _, err := os.Stat(searchPath); err == nil {
			return initIotaSDK(searchPath)
		}
	}

	log.Fatalf("Could not find %v", getIotaSDKLibName())

	return nil
}

var loadedWallet wallets.Wallet

func Load() wallets.Wallet {
	walletProvider := GetWalletProvider()

	if loadedWallet == nil {
		switch walletProvider {
		case ProviderKeyChain:
			loadedWallet = providers.LoadKeyChain(AddressIndex)
		case ProviderLedger:
			loadedWallet = providers.LoadLedgerWallet(getIotaSDK(), AddressIndex)
		case ProviderStronghold:
			loadedWallet = providers.LoadStrongholdWallet(getIotaSDK(), AddressIndex)
		}
	}

	return loadedWallet
}

func InitWallet() {
	walletProvider := GetWalletProvider()

	switch walletProvider {
	case ProviderKeyChain:
		providers.CreateKeyChain()
	case ProviderLedger:
		log.Printf("Ledger wallet provider selected, no initialization required")
	case ProviderStronghold:
		providers.CreateNewStrongholdWallet(getIotaSDK())
	}
}

func Migrate(provider WalletProvider) {
	seedHex := config.GetSeedForMigration()
	if seedHex == "" {
		fmt.Println("No seed to migrate found.")
		return
	}

	seedBytes, err := iotago.DecodeHex(seedHex)
	log.Check(err)
	seed := cryptolib.SeedFromBytes(seedBytes)

	switch provider {
	case ProviderKeyChain:
		providers.MigrateKeyChain(seed)
	case ProviderLedger:
		log.Printf("Ledger wallet provider selected, no migration available")
	case ProviderStronghold:
		providers.MigrateToStrongholdWallet(getIotaSDK(), seed)
	}
}
