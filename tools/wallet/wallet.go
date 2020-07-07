// wallet is a CLI tool for Goshimmer's cryptocurrency wallet, allowing to store the seed in a
// json file, and later generate private and public keys.
//
// Create a new wallet (creates wallet.json):
//   wallet init
// Show private key + public key + account address for index 0 (index optional, default 0):
//   wallet address -n 0
// Query Goshimmer for account balance
//   wallet balance [-n index]
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/wallet"
	nodeapi "github.com/iotaledger/goshimmer/dapps/waspconn/packages/apilib"
)

const goshimmerApi = "127.0.0.1:8080"

type walletConfig struct {
	Seed []byte
}

func check(err error) {
	if err != nil {
		fmt.Printf("[wallet] error: %s\n", err)
		os.Exit(1)
	}
}

func usage(globalFlags *flag.FlagSet) {
	fmt.Printf("Usage: %s [options] [init|address|balance]\n", os.Args[0])
	globalFlags.PrintDefaults()
	os.Exit(1)
}

func main() {
	globalFlags := flag.NewFlagSet("", flag.ExitOnError)
	walletPath := globalFlags.String("w", "wallet.json", "path to wallet.json")
	globalFlags.Parse(os.Args[1:])

	if globalFlags.NArg() < 1 {
		usage(globalFlags)
	}

	switch globalFlags.Arg(0) {
	case "init":
		initWallet(*walletPath)

	case "address":
		pubFlags := flag.NewFlagSet("address", flag.ExitOnError)
		n := pubFlags.Int("n", 0, "address index")
		pubFlags.Parse(globalFlags.Args()[1:])

		dumpAddress(loadWallet(*walletPath), *n)

	case "balance":
		pubFlags := flag.NewFlagSet("balance", flag.ExitOnError)
		n := pubFlags.Int("n", 0, "address index")
		pubFlags.Parse(globalFlags.Args()[1:])

		dumpBalance(loadWallet(*walletPath), *n)

	default:
		usage(globalFlags)
	}
}

func initWallet(walletPath string) {
	walletConfig := &walletConfig{
		Seed: wallet.New().Seed().Bytes(),
	}

	jsonBytes, err := json.MarshalIndent(walletConfig, "", "  ")
	check(err)

	check(ioutil.WriteFile(walletPath, jsonBytes, 0644))
}

func loadWallet(walletPath string) *wallet.Wallet {
	bytes, err := ioutil.ReadFile(walletPath)
	check(err)

	var wc walletConfig
	check(json.Unmarshal(bytes, &wc))

	return wallet.New(wc.Seed)
}

func dumpAddress(wallet *wallet.Wallet, n int) {
	seed := wallet.Seed()
	kp := seed.KeyPair(uint64(n))
	fmt.Printf("Index %d\n", n)
	fmt.Printf("  Private key: %s\n", kp.PrivateKey)
	fmt.Printf("  Public key:  %s\n", kp.PublicKey)
	fmt.Printf("  Address:     %s\n", seed.Address(uint64(n)))
}

func dumpBalance(wallet *wallet.Wallet, n int) {
	seed := wallet.Seed()
	address := seed.Address(uint64(n))

	r, err := nodeapi.GetAccountOutputs(goshimmerApi, &address)
	check(err)

	fmt.Printf("Index %d\n", n)
	fmt.Printf("  Address: %s\n", address)
	fmt.Printf("  Balances:\n")
	if len(r) == 0 {
		fmt.Printf("    (empty)\n")
	} else {
		for _, balances := range r {
			for _, bal := range balances {
				fmt.Printf("    %d %s\n", bal.Value, bal.Color.String())
			}
		}
	}
}
