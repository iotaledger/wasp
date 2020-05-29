package testplugins

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/packages/waspconn/utxodb"
	"github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/hashing"
)

// BLS addresses

var scAddressesStr = []string{
	"r1u46Yss96QKrEaZMw1hcat1x5v8Wyab2xdkDi1hfFf7",
	"eWCs2eZvEkwFZe9jQKPd3Njv7byy8UkRM2X12kQvTbcC",
	"tdGfa4ihTh8WYE5xvufWgTzNJoLrZS7xLyvhg13tKmKh",
}

var descriptions = []string{
	"1. testing nil program",
	"2. testing increment",
	"3. testing FairRoulette",
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
		scOrigParams[i].ProgramHash = *GetProgramHash(i + 1)
	}
}

func GetProgramHash(scIndex int) *hashing.HashValue {
	return hashing.HashStrings(GetScDescription(scIndex))
}

func GetNodeLocations(_ int) []string {
	return nodeLocations
}

func GetScDescription(scIndex int) string {
	return descriptions[scIndex-1]
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
