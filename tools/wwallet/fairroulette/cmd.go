package fairroulette

import (
	"github.com/iotaledger/wasp/tools/wwallet/config/fr"
)

var commands = map[string]func([]string){
	"set":    fr.Config.HandleSetCmd,
	"admin":  adminCmd,
	"status": statusCmd,
	"bet":    betCmd,
}

func Cmd(args []string) {
	fr.Config.HandleCmd(args, commands)
}
