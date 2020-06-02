package main

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/testapilib"
	"github.com/iotaledger/wasp/plugins/testplugins/testaddresses"
	"os"
	"strconv"
)

const requestorIndex = 5
const targetHost = "127.0.0.1:9090"

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
	fmt.Printf("sending test request to builtin sc #%d, addr %s\n through host %s\n",
		scidx, addr.String(), targetHost)

	reqJson := testapilib.RequestTransactionJson{
		RequestorIndex: 1,
		Requests: []testapilib.RequestBlockJson{
			{addr.String(), requestorIndex, nil},
		},
	}
	resp := testapilib.SendTestRequest(targetHost, &reqJson)
	if resp.Error != "" {
		fmt.Printf("Error: %v\n", resp.Error)
	} else {
		fmt.Printf("Success:\nTxid = %s\nnum requests = %d\n", resp.TxId, resp.NumReq)
	}
}
