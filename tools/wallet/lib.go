package wallet

import (
	"encoding/json"
	"io/ioutil"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/wallet"
)

type WalletConfig struct {
	Seed []byte
}

type Wallet struct {
	*wallet.Wallet
}

func Init(walletPath string) error {
	walletConfig := &WalletConfig{
		Seed: wallet.New().Seed().Bytes(),
	}

	jsonBytes, err := json.MarshalIndent(walletConfig, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(walletPath, jsonBytes, 0644)
}

func Load(walletPath string) (*Wallet, error) {
	bytes, err := ioutil.ReadFile(walletPath)
	if err != nil {
		return nil, err
	}

	var wc WalletConfig
	err = json.Unmarshal(bytes, &wc)
	if err != nil {
		return nil, err
	}

	return &Wallet{wallet.New(wc.Seed)}, nil
}
