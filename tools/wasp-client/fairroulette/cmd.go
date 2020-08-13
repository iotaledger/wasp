package fairroulette

import (
	"fmt"
	"os"
	"strings"

	"github.com/iotaledger/wasp/tools/wasp-client/config"
	"github.com/spf13/pflag"
)

var flags *pflag.FlagSet

func HookFlags() *pflag.FlagSet {
	flags = pflag.NewFlagSet("fr", pflag.ExitOnError)
	adminFlags(flags)
	return flags
}

var commands = map[string]func([]string){
	"set":    setCmd,
	"admin":  adminCmd,
	"status": statusCmd,
	"bet":    betCmd,
}

func setCmd(args []string) {
	if len(args) != 2 {
		fmt.Printf("Usage: %s fr set <key> <value>\n", os.Args[0])
		os.Exit(1)
	}
	config.Set("fr."+args[0], args[1])
}

func usage(flags *pflag.FlagSet) {
	cmdNames := make([]string, 0)
	for k, _ := range commands {
		cmdNames = append(cmdNames, k)
	}

	fmt.Printf("Usage: %s fr [options] [%s]\n", os.Args[0], strings.Join(cmdNames, "|"))
	flags.PrintDefaults()
	os.Exit(1)
}

func Cmd(args []string) {
	if len(args) < 1 {
		usage(flags)
	}
	cmd, ok := commands[args[0]]
	if !ok {
		usage(flags)
	}
	cmd(args[1:])
}
