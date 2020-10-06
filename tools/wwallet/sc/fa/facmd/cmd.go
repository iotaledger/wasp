package facmd

import (
	"fmt"
	"os"

	"github.com/iotaledger/wasp/tools/wwallet/sc/fa"
)

func InitCommands(commands map[string]func([]string)) {
	commands["fa"] = cmd
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
