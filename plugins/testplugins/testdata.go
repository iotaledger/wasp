package testplugins

import (
	"github.com/iotaledger/goshimmer/packages/waspconn/utxodb"
	"github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/plugins/testplugins/testaddresses"
)

var descriptions = []string{
	"0. testing nil program",
	"1. testing increment",
	"2. testing FairRoulette",
}

var (
	scOrigParams  []apilib.NewOriginParams
	nodeLocations = []string{
		"127.0.0.1:4000",
		"127.0.0.1:4001",
		"127.0.0.1:4002",
		"127.0.0.1:4003",
	}
)

func init() {
	scOrigParams = make([]apilib.NewOriginParams, testaddresses.NumAddresses())
	for i := range scOrigParams {
		addr, _ := testaddresses.GetAddress(i)
		scOrigParams[i] = apilib.NewOriginParams{
			Address:      *addr,
			OwnerAddress: utxodb.GetAddress(i + 1),
		}
		scOrigParams[i].ProgramHash = *GetProgramHash(i)
	}
}

func GetProgramHash(scIndex int) *hashing.HashValue {
	return hashing.HashStrings(GetScDescription(scIndex))
}

func GetNodeLocations(_ int) []string {
	return nodeLocations
}

func GetScDescription(scIndex int) string {
	return descriptions[scIndex]
}

func GetOriginParams(scIndex int) *apilib.NewOriginParams {
	return &scOrigParams[scIndex]
}
