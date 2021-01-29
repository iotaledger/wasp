package blob

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/tools/wasp-cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/spf13/pflag"
)

func InitCommands(commands map[string]func([]string), flags *pflag.FlagSet) {
	commands["blob"] = blobCmd
}

var subcmds = map[string]func([]string){
	"put": putBlobCmd,
	"get": getBlobCmd,
	"has": hasBlobCmd,
}

func blobCmd(args []string) {
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

	log.Usage("%s blob [%s]\n", os.Args[0], strings.Join(cmdNames, "|"))
}

func putBlobCmd(args []string) {
	if len(args) != 1 {
		log.Usage("%s blob put <filename>\n", os.Args[0])
	}
	data, err := ioutil.ReadFile(args[0])
	log.Check(err)
	hash, err := config.WaspClient().PutBlob(data)
	log.Check(err)
	log.Printf("Blob uploaded. Hash: %s\n", hash)
}

func getBlobCmd(args []string) {
	if len(args) != 1 {
		log.Usage("%s blob get <hash>\n", os.Args[0])
	}
	hash, err := hashing.HashValueFromBase58(args[0])
	log.Check(err)
	data, err := config.WaspClient().GetBlob(hash)
	log.Check(err)
	_, err = os.Stdout.Write(data)
	log.Check(err)
}

func hasBlobCmd(args []string) {
	if len(args) != 1 {
		log.Usage("%s blob has <hash>\n", os.Args[0])
	}
	hash, err := hashing.HashValueFromBase58(args[0])
	log.Check(err)
	ok, err := config.WaspClient().HasBlob(hash)
	log.Check(err)
	log.Printf("%v", ok)
}
