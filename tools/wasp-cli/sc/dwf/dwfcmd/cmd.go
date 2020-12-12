// +build ignore

package dwfcmd

import (
	"fmt"
	"os"

	"github.com/iotaledger/wasp/tools/wasp-cli/sc/dwf"
)

func InitCommands(commands map[string]func([]string)) {
	commands["dwf"] = cmd
}

var subcmds = map[string]func([]string){
	"set":      dwf.Config.HandleSetCmd,
	"admin":    adminCmd,
	"donate":   donateCmd,
	"withdraw": withdrawCmd,
	"status":   statusCmd,
}

func cmd(args []string) {
	dwf.Config.HandleCmd(args, subcmds)
}
