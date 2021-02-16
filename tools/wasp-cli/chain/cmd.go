package chain

import (
	"os"
	"strings"

	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/spf13/pflag"
)

func InitCommands(commands map[string]func([]string), flags *pflag.FlagSet) {
	commands["chain"] = chainCmd

	fs := pflag.NewFlagSet("chain", pflag.ExitOnError)
	initDeployFlags(fs)
	initUploadFlags(fs)
	initAliasFlags(fs)
	flags.AddFlagSet(fs)
}

var subcmds = map[string]func([]string){
	"list":            listCmd,
	"deploy":          deployCmd,
	"info":            infoCmd,
	"list-contracts":  listContractsCmd,
	"deploy-contract": deployContractCmd,
	"list-accounts":   listAccountsCmd,
	"balance":         balanceCmd,
	"list-blobs":      listBlobsCmd,
	"store-blob":      storeBlobCmd,
	"show-blob":       showBlobCmd,
	"log":             logCmd,
	"post-request":    postRequestCmd,
	"call-view":       callViewCmd,
	"activate":        activateCmd,
	"deactivate":      deactivateCmd,
}

func chainCmd(args []string) {
	if len(args) < 1 {
		usage()
	}
	subcmd, ok := subcmds[args[0]]
	if !ok {
		usage()
	}
	subcmd(args[1:])
}

func usage() {
	cmdNames := make([]string, 0)
	for k := range subcmds {
		cmdNames = append(cmdNames, k)
	}

	log.Usage("%s chain [%s]\n", os.Args[0], strings.Join(cmdNames, "|"))
}
