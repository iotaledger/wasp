package dwfcmd

import (
	"fmt"
	"os"

	"github.com/iotaledger/wasp/tools/wwallet/sc/dwf"
)

var commands = map[string]func([]string){
	"set":   dwf.Config.HandleSetCmd,
	"admin": adminCmd,
}

func Cmd(args []string) {
	dwf.Config.HandleCmd(args, commands)
}

func check(err error) {
	if err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}
}
