// wallet is a CLI tool for Goshimmer's cryptocurrency wallet, allowing to store the seed in a
// json file, and later generate private and public keys.
//
// Create a new wallet (creates wallet.json):
//
//   wallet init
//
// Show private key + public key + account address for index 0 (index optional, default 0):
//
//   wallet address -n 0
//
// Query Goshimmer for account balance:
//
//   wallet balance [-n index]
//
// Transfer `amount` IOTA from the given utxodb addres index to the wallet address at index n:
//
//   wallet transfer [-n index] utxodb-index amount
//
package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	nodeapi "github.com/iotaledger/goshimmer/dapps/waspconn/packages/apilib"
	"github.com/iotaledger/goshimmer/dapps/waspconn/packages/utxodb"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/tools/wallet"
)

const goshimmerApi = "127.0.0.1:8080"

func main() {
	globalFlags := flag.NewFlagSet("", flag.ExitOnError)
	walletPath := globalFlags.String("w", "wallet.json", "path to wallet.json")
	globalFlags.Parse(os.Args[1:])

	if globalFlags.NArg() < 1 {
		usage(globalFlags)
	}

	if globalFlags.Arg(0) == "init" {
		check(wallet.Init(*walletPath))
		return
	}

	wallet, err := wallet.Load(*walletPath)
	check(err)

	switch globalFlags.Arg(0) {
	case "address":
		flags := flag.NewFlagSet("address", flag.ExitOnError)
		n := flags.Int("n", 0, "address index")
		flags.Parse(globalFlags.Args()[1:])

		dumpAddress(wallet, *n)

	case "balance":
		flags := flag.NewFlagSet("balance", flag.ExitOnError)
		n := flags.Int("n", 0, "address index")
		flags.Parse(globalFlags.Args()[1:])

		dumpBalance(wallet, *n)

	case "transfer":
		flags := flag.NewFlagSet("transfer", flag.ExitOnError)
		n := flags.Int("n", 0, "address index")
		flags.Parse(globalFlags.Args()[1:])

		if flags.NArg() < 2 {
			transferUsage(globalFlags, flags)
		}

		utxodbIndex, err := strconv.Atoi(flags.Arg(0))
		check(err)
		amount, err := strconv.Atoi(flags.Arg(1))
		check(err)

		transfer(wallet, *n, utxodbIndex, amount)

	default:
		usage(globalFlags)
	}
}

func check(err error) {
	if err != nil {
		fmt.Printf("[wallet] error: %s\n", err)
		os.Exit(1)
	}
}

func usage(globalFlags *flag.FlagSet) {
	fmt.Printf("Usage: %s [options] [init|address|balance|transfer]\n", os.Args[0])
	globalFlags.PrintDefaults()
	os.Exit(1)
}

func transferUsage(globalFlags *flag.FlagSet, flags *flag.FlagSet) {
	fmt.Printf("Usage: %s [options] transfer [-n index] <utxodb-index> <amount>\n", os.Args[0])
	globalFlags.PrintDefaults()
	flags.PrintDefaults()
	os.Exit(1)
}

func dumpAddress(wallet *wallet.Wallet, n int) {
	seed := wallet.Seed()
	kp := seed.KeyPair(uint64(n))
	fmt.Printf("Account index %d\n", n)
	fmt.Printf("  Private key: %s\n", kp.PrivateKey)
	fmt.Printf("  Public key:  %s\n", kp.PublicKey)
	fmt.Printf("  Address:     %s\n", seed.Address(uint64(n)))
}

func dumpBalance(wallet *wallet.Wallet, n int) {
	seed := wallet.Seed()
	address := seed.Address(uint64(n))

	r, err := nodeapi.GetAccountOutputs(goshimmerApi, &address)
	check(err)

	byColor, total := util.OutputBalancesByColor(r)

	fmt.Printf("Account index %d\n", n)
	fmt.Printf("  Address: %s\n", address)
	fmt.Printf("  Balance:\n")
	for color, value := range byColor {
		fmt.Printf("    %s: %d\n", color.String(), value)
	}
	fmt.Printf("    ------\n")
	fmt.Printf("    Total: %d\n", total)
}

func transfer(wallet *wallet.Wallet, n int, utxodbIndex int, amount int) {
	seed := wallet.Seed()
	walletAddress := seed.Address(uint64(n))

	check(nodeapi.PostTransaction(goshimmerApi, makeTransferTx(walletAddress, utxodbIndex, int64(amount))))
}

func makeTransferTx(target address.Address, utxodbIndex int, amount int64) *transaction.Transaction {
	source := utxodb.GetAddress(utxodbIndex)
	sourceOutputs, err := nodeapi.GetAccountOutputs(goshimmerApi, &source)
	check(err)

	oids := make([]transaction.OutputID, 0)
	sum := int64(0)
	for oid, bals := range sourceOutputs {
		containsIotas := false
		for _, b := range bals {
			if b.Color == balance.ColorIOTA {
				sum += b.Value
				containsIotas = true
			}
		}
		if containsIotas {
			oids = append(oids, oid)
		}
		if sum >= amount {
			break
		}
	}

	if sum < amount {
		panic(fmt.Errorf("not enough input balance"))
	}

	inputs := transaction.NewInputs(oids...)

	out := make(map[address.Address][]*balance.Balance)
	out[target] = []*balance.Balance{balance.New(balance.ColorIOTA, amount)}
	if sum > amount {
		out[source] = []*balance.Balance{balance.New(balance.ColorIOTA, sum-amount)}
	}
	outputs := transaction.NewOutputs(out)

	tx := transaction.New(inputs, outputs)
	tx.Sign(utxodb.GetSigScheme(source))
	return tx
}
