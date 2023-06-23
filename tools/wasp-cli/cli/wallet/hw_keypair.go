package wallet

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/usbwallet"

	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

type HWKeyPair struct {
}

func NewHWKeyPair() {
	hub, err := usbwallet.NewLedgerHub()
	log.Check(err)

	wallets := hub.Wallets()
	if len(wallets) == 0 {
		log.Fatal("No ledger device found")
	}

	wallet := wallets[0]
	fmt.Println(wallets)

	err = wallet.Open("")
	log.Check(err)

	status, err := wallet.Status()
	fmt.Println(status)

	accounts := wallets[0].Accounts()
	fmt.Println(accounts)
}
