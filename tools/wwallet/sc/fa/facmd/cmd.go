package facmd

import (
	"fmt"
	"os"

	"github.com/iotaledger/wasp/tools/wwallet/sc/fa"
	"github.com/spf13/pflag"
)

func InitCommands(commands map[string]func([]string), flags *pflag.FlagSet) {
	commands["fa"] = cmd
	flags.AddFlagSet(fa.Config.HookFlags())
}

var subcmds = map[string]func([]string){
	"set":           fa.Config.HandleSetCmd,
	"admin":         adminCmd,
	"status":        statusCmd,
	"start-auction": startAuctionCmd,
	"place-bid":     placeBidCmd,
}

func cmd(args []string) {
	fa.Config.HandleCmd(args, subcmds)
}

func check(err error) {
	if err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}
}
