package wallet

import (
	"fmt"
	"os"

	"github.com/spf13/pflag"
)

func InitCommands(commands map[string]func([]string), flags *pflag.FlagSet) {
	commands["init"] = initCmd
	commands["address"] = addressCmd
	commands["balance"] = balanceCmd
	commands["mint"] = mintCmd
	commands["send-funds"] = sendFundsCmd
	commands["request-funds"] = requestFundsCmd

	fs := pflag.NewFlagSet("wallet", pflag.ExitOnError)
	fs.IntVarP(&addressIndex, "address-index", "i", 0, "address index")
	flags.AddFlagSet(fs)
}

func check(err error) {
	if err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}
}
