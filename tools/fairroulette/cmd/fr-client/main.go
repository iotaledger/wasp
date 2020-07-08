// fr-client allows to use the FairRoulette smart contract from the command line
package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/wasp/tools/fairroulette"
	"github.com/iotaledger/wasp/tools/wallet"
)

const waspApi = "127.0.0.1:9090"
const goshimmerApi = "127.0.0.1:8080"

func main() {
	globalFlags := flag.NewFlagSet("", flag.ExitOnError)
	scAddress := globalFlags.String("sc", "", "SC Address")
	walletPath := globalFlags.String("w", "wallet.json", "path to wallet.json")
	addressIndex := globalFlags.Int("n", 0, "address index")
	globalFlags.Parse(os.Args[1:])

	if globalFlags.NArg() < 1 || len(*scAddress) == 0 {
		usage(globalFlags)
	}

	switch globalFlags.Arg(0) {

	case "state":
		check(fairroulette.DumpState(waspApi, *scAddress))

	case "bet":
		flags := flag.NewFlagSet("bet", flag.ExitOnError)
		flags.Parse(globalFlags.Args()[1:])

		if flags.NArg() < 2 {
			betUsage(globalFlags, flags)
		}

		color, err := strconv.Atoi(flags.Arg(0))
		check(err)
		amount, err := strconv.Atoi(flags.Arg(1))
		check(err)

		wallet, err := wallet.Load(*walletPath)
		check(err)

		scAddress, err := address.FromBase58(*scAddress)
		check(err)

		sigScheme := signaturescheme.ED25519(*wallet.Seed().KeyPair(uint64(*addressIndex)))

		check(fairroulette.PlaceBet(goshimmerApi, scAddress, color, amount, sigScheme))

	default:
		usage(globalFlags)
	}
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func usage(globalFlags *flag.FlagSet) {
	fmt.Printf("Usage: %s [options] [state]\n", os.Args[0])
	globalFlags.PrintDefaults()
	os.Exit(1)
}

func betUsage(globalFlags *flag.FlagSet, flags *flag.FlagSet) {
	fmt.Printf("Usage: %s [options] bet <color> <amount>\n", os.Args[0])
	globalFlags.PrintDefaults()
	flags.PrintDefaults()
	os.Exit(1)
}
