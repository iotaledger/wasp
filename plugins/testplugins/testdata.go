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
	"exZup69X1XwRNHiWWjoYy75aPNgC22YKkPV7sUJSBYA9",
	"dV9hfYyHq7uiCKdKYQoLqyiwX6tN448GRm8UgFpUC3Vo",
	"eiMbhrJjajqnCLmVJqFXzFsh1ZsbCAnJ9wauU8cP8uxL",
}

var (
	scAddresses  []address.Address
	scOrigParams []apilib.NewOriginParams
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

func GetScDescription(scIndex int) string {
	return fmt.Sprintf("Test smart contract #%d", scIndex)
}

// index 1 to 3
func GetScAddress(scIndex int) address.Address {
	return scAddresses[scIndex-1]
}

func GetOriginParams(scIndex int) *apilib.NewOriginParams {
	return &scOrigParams[scIndex-1]
}
