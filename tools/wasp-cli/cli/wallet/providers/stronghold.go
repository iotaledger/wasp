package providers

import (
	"fmt"
	"os"
	"path"
	"strings"
	"syscall"

	"github.com/awnumar/memguard"
	"github.com/tyler-smith/go-bip39"
	"golang.org/x/term"

	walletsdk "github.com/iotaledger/wasp-wallet-sdk"
	"github.com/iotaledger/wasp-wallet-sdk/types"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/keychain"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/wallet/wallets"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

func strongholdStorePath() string {
	return path.Join(config.BaseDir, "client.stronghold")
}

func strongholdStoreExists() bool {
	_, err := os.Stat(strongholdStorePath())
	return err == nil
}

func configureStronghold(sdk *walletsdk.IOTASDK, unlockPassword *memguard.Enclave) (*walletsdk.SecretManager, error) {
	secretManager, err := walletsdk.NewStrongholdSecretManager(sdk, unlockPassword, strongholdStorePath())
	if err != nil {
		return nil, err
	}

	return secretManager, nil
}

func LoadStrongholdWallet(sdk *walletsdk.IOTASDK, addressIndex uint32) wallets.Wallet {
	fmt.Println("Load stronghold wallet")
	password, err := keychain.GetStrongholdPassword()
	log.Check(err)

	secretManager, err := configureStronghold(sdk, password)
	log.Check(err)

	return wallets.NewExternalWallet(secretManager, addressIndex, string(parameters.L1().Protocol.Bech32HRP), types.CoinTypeSMR)
}

func printMnemonic(mnemonic *memguard.Enclave) {
	buffer, err := mnemonic.Open()
	log.Check(err)
	defer buffer.Destroy()
	mnemonicParts := strings.Split(buffer.String(), " ")

	for i, part := range mnemonicParts {
		log.Printf("%s ", part)
		if (i+1)%4 == 0 {
			log.Printf("\n")
		}
	}
}

func readPasswordFromStdin() *memguard.Enclave {
	log.Printf("Password: ")
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin)) //nolint:nolintlint,unconvert // int cast is needed for windows
	log.Check(err)
	return memguard.NewEnclave(passwordBytes)
}

func MigrateToStrongholdWallet(sdk *walletsdk.IOTASDK, seed cryptolib.Seed) {
	log.Printf("Migrating existing seed into Stronghold store.\n\n")

	if strongholdStoreExists() {
		log.Printf("There is an existing Stronghold store in '%s'\nIt will be overwritten once you enter a password.\n\n", strongholdStorePath())
	}

	log.Printf("Enter a secure password.\n")
	unlockPassword := readPasswordFromStdin()
	log.Printf("\n")

	s := seed.SubSeed(0)
	mnemonicStr, err := bip39.NewMnemonic(s[:])
	log.Check(err)
	mnemonic := memguard.NewEnclave([]byte(mnemonicStr))

	createNewStrongholdWallet(sdk, mnemonic, unlockPassword)
}

func createNewStrongholdWallet(sdk *walletsdk.IOTASDK, mnemonic *memguard.Enclave, password *memguard.Enclave) {
	if strongholdStoreExists() {
		err := os.Remove(strongholdStorePath())
		log.Check(err)
	}

	log.Printf("\n\n")

	secretManager, err := configureStronghold(sdk, password)
	log.Check(err)

	success, err := secretManager.StoreMnemonic(mnemonic)
	log.Check(err)

	if success {
		log.Printf("Stronghold store generated.\nWrite down the following mnemonic to recover your seed at a later time.\n\n")
		printMnemonic(mnemonic)
	} else {
		log.Printf("Setting the mnemonic failed unexpectedly.")
		return
	}

	err = keychain.SetStrongholdPassword(password)
	log.Check(err)
}

func CreateNewStrongholdWallet(sdk *walletsdk.IOTASDK) {
	log.Printf("Creating new Stronghold store.\n\n")

	if strongholdStoreExists() {
		log.Printf("There is an existing Stronghold store in '%s'\nIt will be overwritten once you enter a password.\n\n", strongholdStorePath())
	}

	log.Printf("Enter a secure password.\n")
	unlockPassword := readPasswordFromStdin()
	log.Printf("\n")

	mnemonic, err := sdk.Utils().GenerateMnemonic()
	log.Check(err)

	createNewStrongholdWallet(sdk, mnemonic, unlockPassword)
}
