package providers

import (
	"os"
	"path"
	"strings"
	"syscall"

	"golang.org/x/term"

	walletsdk "github.com/iotaledger/wasp-wallet-sdk"
	"github.com/iotaledger/wasp-wallet-sdk/types"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/wallet/wallets"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

func strongholdStorePath() string {
	homeDir, err := os.UserHomeDir()
	log.Check(err)

	return path.Join(homeDir, ".config/wasp-cli", "client.stronghold")
}

func strongholdStoreExists() bool {
	_, err := os.Stat(strongholdStorePath())
	return err == nil
}

func configureStronghold(sdk *walletsdk.IOTASDK, unlockPassword string) (*walletsdk.SecretManager, error) {
	secretManager, err := walletsdk.NewStrongholdSecretManager(sdk, types.StrongholdSecretManagerStronghold{
		SnapshotPath: strongholdStorePath(),
		Password:     unlockPassword,
	})

	if err != nil {
		return nil, err
	}

	return secretManager, nil
}

func LoadStrongholdWallet(sdk *walletsdk.IOTASDK, addressIndex uint32) wallets.Wallet {
	keyChain := NewKeyChain()
	password, err := keyChain.GetStrongholdPassword()
	log.Check(err)

	secretManager, err := configureStronghold(sdk, password)
	log.Check(err)

	return wallets.NewExternalWallet(secretManager, addressIndex, string(parameters.L1().Protocol.Bech32HRP), types.CoinTypeSMR)
}

func printMnemonic(mnemonic string) {
	mnemonicParts := strings.Split(mnemonic, " ")

	for i, part := range mnemonicParts {
		log.Printf("%s ", part)
		if (i+1)%4 == 0 {
			log.Printf("\n")
		}
	}
}

func readPasswordFromStdin() string {
	log.Printf("Password: ")
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin)) //nolint:nolintlint,unconvert // int cast is needed for windows
	if err != nil {
		panic(err)
	}
	return string(passwordBytes)
}

func CreateNewStrongholdWallet(sdk *walletsdk.IOTASDK) {
	log.Printf("Creating new Stronghold store.\n\n")

	if strongholdStoreExists() {
		log.Printf("There is an existing Stronghold store in '%s'\nIt will be overwritten once you enter a password.\n\n", strongholdStorePath())
	}

	log.Printf("Enter a secure password.\n")
	unlockPassword := readPasswordFromStdin()
	log.Printf("\n")

	if strongholdStoreExists() {
		err := os.Remove(strongholdStorePath())
		log.Check(err)
	}

	log.Printf("\n\n")

	secretManager, err := configureStronghold(sdk, unlockPassword)
	log.Check(err)

	mnemonic, err := sdk.Utils().GenerateMnemonic()
	log.Check(err)

	success, err := secretManager.StoreMnemonic(*mnemonic)
	log.Check(err)

	if success {
		log.Printf("Stronghold store generated.\nWrite down the following mnemonic to recover your seed at a later time.\n\n")
		printMnemonic(*mnemonic)
	} else {
		log.Printf("Setting the mnemonic failed unexpectedly.")
		return
	}

	keyChain := NewKeyChain()
	err = keyChain.SetStrongholdPassword(unlockPassword)
	log.Check(err)
}
