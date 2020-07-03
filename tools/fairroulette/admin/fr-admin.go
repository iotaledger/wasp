// fr-admin allows to control the FairRoulette smart contract from the command line
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	nodeapi "github.com/iotaledger/goshimmer/dapps/waspconn/packages/apilib"
	"github.com/iotaledger/goshimmer/dapps/waspconn/packages/utxodb"
	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/vm/examples/fairroulette"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func usage(globalFlags *flag.FlagSet) {
	fmt.Printf("Usage: %s [options] [init]\n", os.Args[0])
	globalFlags.PrintDefaults()
	os.Exit(1)
}

type scConfig struct {
	description  string
	quorum       int
	ownerAddress address.Address
	programHash  string
}

type globalConfig struct {
	waspCommitteeApi     []string
	waspCommitteePeering []string
	goshimmerHost        string
	sc                   scConfig
}

var config = globalConfig{
	waspCommitteeApi: []string{
		"127.0.0.1:9090",
		"127.0.0.1:9091",
		"127.0.0.1:9092",
		"127.0.0.1:9093",
	},
	waspCommitteePeering: []string{
		"127.0.0.1:4000",
		"127.0.0.1:4001",
		"127.0.0.1:4002",
		"127.0.0.1:4003",
	},
	goshimmerHost: "127.0.0.1:8080",
	sc: scConfig{
		description:  "FairRoulette smart contract",
		quorum:       3,
		ownerAddress: utxodb.GetAddress(1),
		programHash:  fairroulette.ProgramHash,
	},
}

func main() {
	globalFlags := flag.NewFlagSet("", flag.ExitOnError)
	globalFlags.Parse(os.Args[1:])

	if globalFlags.NArg() < 1 {
		usage(globalFlags)
	}

	switch globalFlags.Arg(0) {
	case "init":
		initSC()
	default:
		usage(globalFlags)
	}
}

func initSC() {
	scAddress := genDKSets()
	origTx := createOriginTx(scAddress)
	color := balance.Color(origTx.ID())
	putScData(scAddress, &color)
	activateSC(scAddress)
	postOriginTx(origTx)
	fmt.Printf("Initialized %s\n", config.sc.description)
	fmt.Printf("SC Address: %s\n", scAddress.String())
}

func genDKSets() *address.Address {
	scAddress, err := waspapi.GenerateNewDistributedKeySet(
		config.waspCommitteeApi,
		uint16(len(config.waspCommitteeApi)),
		uint16(config.sc.quorum),
	)
	check(err)
	return scAddress
}

func putScData(scAddress *address.Address, color *balance.Color) {
	bootupData := registry.BootupData{
		Address:        *scAddress,
		Color:          *color,
		OwnerAddress:   config.sc.ownerAddress,
		CommitteeNodes: config.waspCommitteePeering,
		AccessNodes:    []string{},
	}
	for _, host := range config.waspCommitteeApi {
		check(waspapi.PutSCData(host, bootupData))
	}
}

func createOriginTx(scAddress *address.Address) *sctransaction.Transaction {
	origTx, err := waspapi.CreateOriginUtxodb(waspapi.CreateOriginParams{
		Address:              *scAddress,
		OwnerSignatureScheme: utxodb.GetSigScheme(config.sc.ownerAddress),
		ProgramHash:          progHash(),
		Variables: kv.FromGoMap(map[kv.Key][]byte{
			"description": []byte(config.sc.description),
		}),
	})
	check(err)
	return origTx
}

func progHash() hashing.HashValue {
	hash, err := hashing.HashValueFromBase58(config.sc.programHash)
	check(err)
	return hash
}

func activateSC(scAddress *address.Address) {
	for _, host := range config.waspCommitteeApi {
		check(waspapi.ActivateSC(host, scAddress.String()))
	}
}

func postOriginTx(origTx *sctransaction.Transaction) {
	check(nodeapi.PostTransaction(config.goshimmerHost, origTx.Transaction))
}
