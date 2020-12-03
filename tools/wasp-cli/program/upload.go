package program

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/tools/wasp-cli/config"
)

func uploadCmd(args []string) {
	if len(args) != 4 {
		uploadUsage()
	}

	code, err := ioutil.ReadFile(args[0])
	check(err)
	vmtype := args[1]
	description := args[2]
	nodes := parseIntList(args[3])

	for _, host := range config.CommitteeApi(nodes) {
		hash, err := client.NewWaspClient(host).PutProgram(vmtype, description, code)
		check(err)

		fmt.Printf("Program uploaded to host %s. Program hash: %s\n", host, hash.String())
	}
}

func uploadUsage() {
	fmt.Printf("Usage: %s program upload <filename> <vmtype> <description> <nodes>\n", os.Args[0])
	fmt.Printf("Example: %s program upload program-code.bin wasm 'Example smart contract' '0,1,2,3'\n", os.Args[0])
	os.Exit(1)
}
