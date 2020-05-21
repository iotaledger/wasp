package testplugins

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/packages/waspconn/utxodb"
	"github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/hashing"
)

// BLS addresses
const (
	blsaddr1 = "exZup69X1XwRNHiWWjoYy75aPNgC22YKkPV7sUJSBYA9"
	blsaddr2 = "dV9hfYyHq7uiCKdKYQoLqyiwX6tN448GRm8UgFpUC3Vo"
	blsaddr3 = "eiMbhrJjajqnCLmVJqFXzFsh1ZsbCAnJ9wauU8cP8uxL"
)

var (
	addr1 address.Address
	addr2 address.Address
	addr3 address.Address
	SC1   apilib.NewOriginParams
	SC2   apilib.NewOriginParams
	SC3   apilib.NewOriginParams
)

func init() {
	var err error

	addr1, err = address.FromBase58(blsaddr1)
	if err != nil {
		panic(err)
	}
	addr2, err = address.FromBase58(blsaddr2)
	if err != nil {
		panic(err)
	}
	addr3, err = address.FromBase58(blsaddr3)
	if err != nil {
		panic(err)
	}
	ownerAddress := utxodb.GetAddress(1)
	SC1 = apilib.NewOriginParams{
		Address:      addr1,
		OwnerAddress: ownerAddress,
		Description:  "Test smart contract 1 one",
	}
	SC1.ProgramHash = *hashing.HashStrings(SC1.Description)

	ownerAddress = utxodb.GetAddress(2)
	SC2 = apilib.NewOriginParams{
		Address:      addr2,
		OwnerAddress: ownerAddress,
		Description:  "Test smart contract 2 two",
	}
	SC2.ProgramHash = *hashing.HashStrings(SC2.Description)

	ownerAddress = utxodb.GetAddress(3)
	SC3 = apilib.NewOriginParams{
		Address:      addr3,
		OwnerAddress: ownerAddress,
		Description:  "Test smart contract 3 three",
	}
	SC3.ProgramHash = *hashing.HashStrings(SC3.Description)
}
