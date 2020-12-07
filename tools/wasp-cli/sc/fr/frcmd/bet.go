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
	log.Check(err)
	amount, err := strconv.Atoi(args[1])
	log.Check(err)

	_, err = fr.Client().Bet(color, amount)
	log.Check(err)
}
