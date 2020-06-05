package main

import (
	"fmt"
	"github.com/iotaledger/goshimmer/packages/waspconn/utxodb"
	"github.com/iotaledger/wasp/packages/testapilib"
	"github.com/iotaledger/wasp/plugins/testplugins/testaddresses"
	"os"
	"strconv"
)

const requestorIndex = 5
const targetHost = "127.0.0.1:8080"

func main() {
	fmt.Println("--------------------------------------------------")
	if len(os.Args) != 2 {
		fmt.Printf("usage: 'sendreq <sc index>' where <sc index> = 0 .. %d\n", testaddresses.NumAddresses()-1)
		os.Exit(1)
	}
	scidx, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Printf("%v\n", err)
		fmt.Printf("usage: 'sendreq <sc index>' where <sc index> = 0 .. %d\n", testaddresses.NumAddresses()-1)
		os.Exit(1)
	} else {
		if scidx < 0 || scidx > testaddresses.NumAddresses()-1 {
			fmt.Printf("usage: 'sendreq <sc index>' where <sc index> = 0 .. %d\n", testaddresses.NumAddresses()-1)
			os.Exit(1)
		}
	}
	addr, enabled := testaddresses.GetAddress(scidx)
	if !enabled {
		fmt.Printf("address disabled\n")
		os.Exit(0)
	}

	senderAddr := utxodb.GetAddress(requestorIndex)
	fmt.Printf("sending test request to sc addr #%d, addr %s\n through host %s from sender addr %s\n",
		scidx, addr.String(), targetHost, senderAddr.String())

	reqJson := &testapilib.RequestBlockJson{addr.String(), 0, 100, nil}

	tx, err := testapilib.PostRequestTransaction(targetHost, &senderAddr, []*testapilib.RequestBlockJson{reqJson})

	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Success:\nTxid = %s\n", tx.ID().String())
	}
}
