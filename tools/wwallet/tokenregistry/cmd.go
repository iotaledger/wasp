package tokenregistry

import (
	"fmt"
	"os"

	"github.com/iotaledger/wasp/tools/wwallet/config/tr"
)

var commands = map[string]func([]string){
	"set":    tr.Config.HandleSetCmd,
	"admin":  adminCmd,
	"status": statusCmd,
	"query":  queryCmd,
	"mint":   mintCmd,
}

func Cmd(args []string) {
	tr.Config.HandleCmd(args, commands)
}

func check(err error) {
	if err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}
}
