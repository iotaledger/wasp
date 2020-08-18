package fairroulette

import (
	"fmt"
	"os"
	"strconv"

	"github.com/iotaledger/wasp/tools/wasp-client/scclients"
)

func betCmd(args []string) {
	if len(args) != 2 {
		scConfig.PrintUsage("bet <color> <amount>")
		os.Exit(1)
	}

	color, err := strconv.Atoi(args[0])
	check(err)
	amount, err := strconv.Atoi(args[1])
	check(err)

	check(scclients.GetFRClient().Bet(color, amount))
}

func check(err error) {
	if err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}
}
