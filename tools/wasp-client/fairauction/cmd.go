package fairauction

import (
	"fmt"
	"os"

	"github.com/iotaledger/wasp/tools/wasp-client/config/fa"
)

var commands = map[string]func([]string){
	"set":           fa.Config.HandleSetCmd,
	"admin":         adminCmd,
	"status":        statusCmd,
	"start-auction": startAuctionCmd,
	"place-bid":     placeBidCmd,
}

func Cmd(args []string) {
	fa.Config.HandleCmd(args, commands)
}

func check(err error) {
	if err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}
}
