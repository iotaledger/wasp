// wallet is a CLI tool for Goshimmer's cryptocurrency wallet, allowing to store the seed in a
// json file, and later generate private and public keys.
//
// Create a new wallet (creates wallet.json):
//
//   wallet init
//
// Show private key + public key + account address for index 0 (index optional, default 0):
//
//   wallet address -i 0
//
// Query Goshimmer for account balance:
//
//   wallet balance [-i index]
//
// Transfer `amount` IOTA from the given utxodb addres index to the wallet address at index n:
//
//   wallet transfer [-i index] utxodb-index amount
//
package wallet

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/pflag"
)

func HookFlags() *pflag.FlagSet {
	flags := pflag.NewFlagSet("wallet", pflag.ExitOnError)
	flags.IntVarP(&addressIndex, "address-index", "i", 0, "address index")
	return flags
}

func Cmd(args []string) {
	if len(args) == 0 {
		usage()
	}

	if args[0] == "init" {
		check(Init())
		return
	}

	switch args[0] {
	case "address":
		dumpAddress()

	case "balance":
		dumpBalance()

	case "transfer":
		if len(args) != 3 {
			fmt.Printf("Usage: %s wallet transfer <utxodb-index> <amount>\n", os.Args[0])
			os.Exit(1)
		}

		utxodbIndex, err := strconv.Atoi(args[1])
		check(err)
		amount, err := strconv.Atoi(args[2])
		check(err)

		transfer(utxodbIndex, amount)

	default:
		usage()
	}
}

func check(err error) {
	if err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Printf("Usage: %s wallet [init|address|balance|transfer]\n", os.Args[0])
	os.Exit(1)
}
