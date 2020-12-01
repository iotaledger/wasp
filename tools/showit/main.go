package main

import (
	"fmt"
	"github.com/iotaledger/goshimmer/client"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"os"
)

const defaultGoshimmer = "waspdev02.iota.cafe:8080"

func main() {
	if len(os.Args) < 3 {
		fmt.Printf("usage: showit <tx|addr> <addr or txid> [<goshimmer host>]\n")
		os.Exit(1)
	}
	node := defaultGoshimmer
	if len(os.Args) >= 4 {
		node = os.Args[3]
	}
	fmt.Printf("using %s\n", node)

	switch os.Args[1] {
	case "tx":
		PrintTransactionById(node, os.Args[2])

	case "addr":
		PrintAddress(node, os.Args[2])
	}

}

func PrintTransactionById(node string, txidBase58 string) {
	gsclient := client.NewGoShimmerAPI("http://" + node)
	resp, err := gsclient.GetTransactionByID(txidBase58)
	if err != nil {
		fmt.Printf("GetTransactionByID: error while querying transaction %s: %v", txidBase58, err)
		return
	}

	fmt.Printf("-- Transaction: %s\n", txidBase58)
	fmt.Printf("-- Data payload: %d bytes\n", len(resp.Transaction.DataPayload))
	fmt.Printf("-- Inputs:\n")
	for _, inp := range resp.Transaction.Inputs {
		fmt.Printf("    %s\n", inp)
	}
	fmt.Printf("-- Outputs:\n")
	for _, outp := range resp.Transaction.Outputs {
		fmt.Printf("    Target: %s\n", outp.Address)
		for _, bal := range outp.Balances {
			fmt.Printf("        %s: %d\n", bal.Color, bal.Value)
		}
	}
	fmt.Printf("-- Inclusion state:\n    %+v\n", resp.InclusionState)
	fmt.Printf("-- Error:\n%s\n", resp.Error)
}

func PrintAddress(node string, addrBase58 string) {
	gsclient := client.NewGoShimmerAPI("http://" + node)
	resp, err := gsclient.GetUnspentOutputs([]string{addrBase58})
	if err != nil {
		fmt.Printf("GetUnspentOutputs: error while querying UTXOs for %s: %v", addrBase58, err)
		return
	}
	fmt.Printf("-- UTXOs of address %s\n", addrBase58)
	for _, a := range resp.UnspentOutputs {
		if a.Address != addrBase58 {
			continue
		}
		for _, out := range a.OutputIDs {
			fmt.Printf("     OutputId: %s\n", out.ID)
			oid, err := transaction.OutputIDFromBase58(out.ID)
			if err != nil {
				fmt.Printf("        error: %v\n", err)
				continue
			}
			fmt.Printf("        Addr: %s, Tx: %s\n", oid.Address().String(), oid.TransactionID().String())
			fmt.Printf("        Inclusion state: %+v\n", out.InclusionState)
			for _, bal := range out.Balances {
				fmt.Printf("           %s: %d\n", bal.Color, bal.Value)
			}
		}
	}
	fmt.Printf("-- Error:\n%s\n", resp.Error)
}
