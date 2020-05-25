package testplugins

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/packages/waspconn/utxodb"
	"github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/hashing"
)

// BLS addresses

var scAddressesStr = []string{
	"mQDimPXfu5auHUksTVAzXPDM1cJr1ryx4vgo1XvPtmpX",
	"ooRmabFjCxgJSW9KszW4gPPBBZFi4ttb7nr5RUpJbWth",
	"kbBfpmYwdxS9f3s1L7u5RqvetJ7MYM4kYy27NzWC561p",
}

var (
	scAddresses   []address.Address
	scOrigParams  []apilib.NewOriginParams
	nodeLocations = []string{
		"127.0.0.1:4000",
		"127.0.0.1:4001",
		"127.0.0.1:4002",
		"127.0.0.1:4003",
	}
)

func init() {
	var err error
	scAddresses = make([]address.Address, len(scAddressesStr))
	scOrigParams = make([]apilib.NewOriginParams, len(scAddressesStr))

	for i := range scAddresses {
		scAddresses[i], err = address.FromBase58(scAddressesStr[i])
		if err != nil {
			panic(err)
		}
	}
	for i := range scAddresses {
		ownerAddress := utxodb.GetAddress(i + 1)
		scOrigParams[i] = apilib.NewOriginParams{
			Address:      scAddresses[i],
			OwnerAddress: ownerAddress,
		}
		scOrigParams[i].ProgramHash = *hashing.HashStrings(GetScDescription(i + 1))
	}
}

func GetNodeLocations(_ int) []string {
	return nodeLocations
}

func GetScDescription(scIndex int) string {
	return fmt.Sprintf("Test smart contract #%d", scIndex)
}

// index 1 to 3
func GetScAddress(scIndex int) address.Address {
	return scAddresses[scIndex-1]
}

func NumBuiltinScAddresses() int {
	return len(scAddresses)
}

func GetOriginParams(scIndex int) *apilib.NewOriginParams {
	return &scOrigParams[scIndex-1]
}
