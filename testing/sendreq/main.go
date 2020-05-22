package main

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/testapilib"
	"github.com/iotaledger/wasp/plugins/testplugins"
)

const requestorIndex = 5

func main() {
	reqJson := testapilib.RequestTransactionJson{
		RequestorIndex: 1,
		Requests: []testapilib.RequestBlockJson{
			{testplugins.GetScAddress(1).String(), requestorIndex, nil},
		},
	}
	err := testapilib.SendTestRequest("127.0.0.1:9090", &reqJson)
	if err != nil {
		fmt.Printf("error: %v\n", err)
	} else {
		fmt.Println("OK")
	}
}
