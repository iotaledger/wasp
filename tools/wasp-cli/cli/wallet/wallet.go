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

type WalletScheme string

const (
	SchemeKeyChain   WalletScheme = "keychain"
	SchemeLedger     WalletScheme = "sdk_ledger"
	SchemeStronghold WalletScheme = "sdk_stronghold"
)

func GetWalletScheme() WalletScheme {
	scheme := WalletScheme(config.GetWalletSchemeString())

	switch scheme {
	case SchemeLedger, SchemeKeyChain, SchemeStronghold:
		return scheme
	}
	return SchemeKeyChain
}

func SetWalletScheme(scheme WalletScheme) error {
	switch scheme {
	case SchemeLedger, SchemeKeyChain, SchemeStronghold:
		config.SetWalletSchemeString(string(scheme))
		return nil
	}
	return errors.New("invalid wallet scheme provided")
}

func getIotaSDKLibName() string {
	switch runtime.GOOS {
	case "windows":
		return "iota_sdk_native.dll"
	case "linux":
		return "libiota_sdk_native.so"
	case "darwin":
		return "libiota_sdk_native.dylib"
	default:
		panic(fmt.Sprintf("unsupported OS: %s", runtime.GOOS))
	}
}

func getIotaSDK() *wasp_wallet_sdk.IOTASDK {
	// LoadLibrary (windows) and dlLoad (linux) have different search path behaviors
	// For now, use a relative path - as it will eventually be shipped with a release.
	ex, err := os.Executable()
	wd := filepath.Dir(ex)
	log.Check(err)

	libPath := path.Join(wd, getIotaSDKLibName())
	sdk, err := wasp_wallet_sdk.NewIotaSDK(libPath)
	log.Check(err)

	_, err = sdk.InitLogger(types.ILoggerConfig{
		LevelFilter: config.GetWalletLogLevel(),
	})
	log.Check(err)

	return sdk
}

var loadedWallet wallets.Wallet

func Load() wallets.Wallet {
	walletScheme := GetWalletScheme()

	if loadedWallet == nil {
		switch walletScheme {
		case SchemeKeyChain:
			loadedWallet = providers.LoadKeyChain(AddressIndex)
		case SchemeLedger:
			loadedWallet = providers.LoadLedgerWallet(getIotaSDK(), AddressIndex)
		case SchemeStronghold:
			loadedWallet = providers.LoadStrongholdWallet(getIotaSDK(), AddressIndex)
		}
	}

	return nil
	seedHex := viper.GetString("wallet.seed")

	useLegacyDerivation := viper.GetBool("wallet.useLegacyDerivation")
	if seedHex == "" {
		log.Fatal("call `init` first")
	}

	masterSeed, err := iotago.DecodeHex(seedHex)
	log.Check(err)

	kp := cryptolib.KeyPairFromSeed(cryptolib.SubSeed(masterSeed, uint32(AddressIndex), useLegacyDerivation))

	return &Wallet{KeyPair: kp, AddressIndex: AddressIndex}
}

func InitWallet() {
	walletScheme := GetWalletScheme()

	switch walletScheme {
	case SchemeKeyChain:
		providers.CreateKeyChain()
	case SchemeLedger:
		log.Printf("Ledger wallet scheme selected, no initialization required")
	case SchemeStronghold:
		providers.CreateNewStrongholdWallet(getIotaSDK())
	}
}

func Migrate(scheme WalletScheme) {
	seedHex := config.GetSeedForMigration()
	if seedHex == "" {
		fmt.Println("No seed to migrate found.")
		return
	}

	seedBytes, err := iotago.DecodeHex(seedHex)
	log.Check(err)
	seed := cryptolib.SeedFromBytes(seedBytes)

	switch scheme {
	case SchemeKeyChain:
		providers.MigrateKeyChain(seed)
	case SchemeLedger:
		log.Printf("Ledger wallet scheme selected, no migration available")
	case SchemeStronghold:
		providers.MigrateToStrongholdWallet(getIotaSDK(), seed)
	}
}
