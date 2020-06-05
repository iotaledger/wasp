package testplugins

import (
	"github.com/iotaledger/goshimmer/packages/waspconn/utxodb"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/sctransaction/origin"
	"github.com/iotaledger/wasp/plugins/testplugins/testaddresses"
)

var descriptions = []string{
	"0. testing nil program",
	"1. testing increment",
	"2. testing FairRoulette",
}

var (
	scOrigParams  []origin.NewOriginParams
	nodeLocations = []string{
		"127.0.0.1:4000",
		"127.0.0.1:4001",
		"127.0.0.1:4002",
		"127.0.0.1:4003",
	}
)

func init() {
	scOrigParams = make([]origin.NewOriginParams, testaddresses.NumAddresses())
	for i := range scOrigParams {
		addr, _ := testaddresses.GetAddress(i)
		scOrigParams[i] = origin.NewOriginParams{
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

func GetOriginParams(scIndex int) *origin.NewOriginParams {
	return &scOrigParams[scIndex]
}
