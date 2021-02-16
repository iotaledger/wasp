// +build ignore

package facmd

import (
	"fmt"
	"os"

	"github.com/iotaledger/wasp/tools/wasp-cli/sc/fa"
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
