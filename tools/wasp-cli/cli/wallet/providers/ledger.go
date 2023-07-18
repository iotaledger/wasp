package providers

import (
	walletsdk "github.com/iotaledger/wasp-wallet-sdk"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/wallet/wallets"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

func LoadLedgerWallet(sdk *walletsdk.IOTASDK, addressIndex uint32) wallets.Wallet {
	secretManager, err := walletsdk.NewLedgerSecretManager(sdk, false)
	log.Check(err)

	status, err := secretManager.GetLedgerStatus()
	log.Check(err)

	if !status.Connected {
		log.Fatalf("Ledger could not be found.")
	}

	if status.Locked {
		log.Fatalf("Ledger is locked")
	}

	hrp := parameters.L1().Protocol.Bech32HRP
	coinType := MapCoinType(hrp)

	return wallets.NewExternalWallet(secretManager, addressIndex, string(hrp), coinType)
}
