package fairauction

import "github.com/iotaledger/wasp/tools/wasp-client/config/fa"

var commands = map[string]func([]string){
	"set":    fa.Config.HandleSetCmd,
	"status": statusCmd,
}

func Cmd(args []string) {
	fa.Config.HandleCmd(args, commands)
}
