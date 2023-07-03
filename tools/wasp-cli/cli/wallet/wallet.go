package wallet

import (
	"errors"
	"fmt"
	"runtime"

	wasp_wallet_sdk "github.com/iotaledger/wasp-wallet-sdk"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/wallet/providers"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/wallet/wallets"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

var AddressIndex uint32

type WalletScheme string

const (
	SchemeInMemory   WalletScheme = "in_memory"
	SchemeLedger     WalletScheme = "sdk_ledger"
	SchemeStronghold WalletScheme = "sdk_stronghold"
)

func GetWalletScheme() WalletScheme {
	scheme := WalletScheme(config.GetWalletSchemeString())

	switch scheme {
	case SchemeLedger, SchemeInMemory, SchemeStronghold:
		return scheme
	}
	return SchemeInMemory
}

func SetWalletScheme(scheme WalletScheme) error {
	switch scheme {
	case SchemeLedger, SchemeInMemory, SchemeStronghold:
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
	sdk, err := wasp_wallet_sdk.NewIotaSDK(getIotaSDKLibName())
	log.Check(err)
	return sdk
}

func Load() wallets.Wallet {
	walletScheme := GetWalletScheme()

	log.Printf("Scheme: %v\n", walletScheme)
	switch walletScheme {
	case SchemeInMemory:
		return providers.LoadInMemory(AddressIndex)
	case SchemeLedger:
		return providers.LoadLedgerWallet(getIotaSDK(), AddressIndex)
	case SchemeStronghold:
		return providers.LoadStrongholdWallet(getIotaSDK(), AddressIndex)
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
	case SchemeInMemory:
		providers.CreateNewInMemory()
	case SchemeLedger:
		log.Printf("Ledger wallet scheme selected, no initialization required")
	case SchemeStronghold:
		providers.CreateNewStrongholdWallet(getIotaSDK())
	}
}
