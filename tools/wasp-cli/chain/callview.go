package chain

import (
	"os"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/tools/wasp-cli/util"
)

func callViewCmd(args []string) {
	if len(args) < 2 {
		log.Fatal("Usage: %s chain call-view <name> <funcname> [params]", os.Args[0])
	}
	r, err := SCClient(coretypes.Hn(args[0])).CallView(args[1], util.EncodeParams(args[2:]))
	log.Check(err)
	util.PrintDictAsJson(r)
}
