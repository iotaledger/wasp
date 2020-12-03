package frcmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/iotaledger/wasp/tools/wasp-cli/sc/fr"
)

func betCmd(args []string) {
	if len(args) != 2 {
		fr.Config.PrintUsage("bet <color> <amount>")
		os.Exit(1)
	}

	color, err := strconv.Atoi(args[0])
	check(err)
	amount, err := strconv.Atoi(args[1])
	check(err)

	_, err = fr.Client().Bet(color, amount)
	check(err)
}

func check(err error) {
	if err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}
}
