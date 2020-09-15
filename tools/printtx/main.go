package main

import (
	"fmt"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/nodeclient/goshimmer"
	"os"
)

const defaultGoshimmer = "waspdev02.iota.cafe:8080"

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("usage: printtx <txid>\n")
		os.Exit(1)
	}
	_, err := valuetransaction.IDFromBase58(os.Args[1])
	if err != nil {
		fmt.Printf("error: %+v\n", err)
		os.Exit(1)
	}
	node := defaultGoshimmer
	if len(os.Args) >= 3 {
		node = os.Args[2]
	}
	fmt.Printf("using GoShimmer %s\n", node)
	gclient := goshimmer.NewGoshimmerClient(node)
	gclient.PrintTransactionById(os.Args[1])
}
