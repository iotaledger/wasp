package admin

import (
	"fmt"
	"os"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	nodeapi "github.com/iotaledger/goshimmer/dapps/waspconn/packages/apilib"
	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/vm/examples/fairroulette"
	"github.com/iotaledger/wasp/tools/fairroulette/config"
	"github.com/iotaledger/wasp/tools/fairroulette/wallet"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const scDescription = "FairRoulette smart contract"
const scProgramHash = fairroulette.ProgramHash

func AdminCmd(args []string) {
	if len(args) < 1 {
		usage()
	}

	switch args[0] {
	case "init":
		flags := pflag.NewFlagSet("init", pflag.ExitOnError)
		quorum := flags.IntP("quorum", "t", 3, "quorum")
		flags.IntSliceP("committee", "n", []int{0, 1, 2, 3}, "committee")
		flags.Parse(args[1:])

		viper.BindPFlag("fairroulette.committee", flags.Lookup("committee"))

		initSC(*quorum)
	default:
		usage()
	}
}

func check(err error) {
	if err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Printf("Usage: %s admin [init]\n", os.Args[0])
	os.Exit(1)
}

func initSC(quorum int) {
	scAddress := genDKSets(quorum)
	origTx := createOriginTx(scAddress)
	color := balance.Color(origTx.ID())
	putScData(scAddress, &color)
	activateSC(scAddress)
	postOriginTx(origTx)
	fmt.Printf("Initialized %s\n", scDescription)
	fmt.Printf("SC Address: %s\n", scAddress.String())
	config.SetSCAddress(scAddress.String())
}

func genDKSets(quorum int) *address.Address {
	scAddress, err := waspapi.GenerateNewDistributedKeySet(
		config.CommitteeApi(committee()),
		uint16(len(committee())),
		uint16(quorum),
	)
	check(err)
	return scAddress
}

func putScData(scAddress *address.Address, color *balance.Color) {
	bootupData := registry.BootupData{
		Address:        *scAddress,
		Color:          *color,
		OwnerAddress:   ownerAddress(),
		CommitteeNodes: config.CommitteePeering(committee()),
		AccessNodes:    []string{},
	}
	for _, host := range config.CommitteeApi(committee()) {
		check(waspapi.PutSCData(host, bootupData))
	}
}

func createOriginTx(scAddress *address.Address) *sctransaction.Transaction {
	origTx, err := waspapi.CreateOrigin(config.GoshimmerApi(), waspapi.CreateOriginParams{
		Address:              *scAddress,
		OwnerSignatureScheme: wallet.Load().SignatureScheme(),
		ProgramHash:          progHash(),
		Variables: kv.FromGoMap(map[kv.Key][]byte{
			"description": []byte(scDescription),
		}),
	})
	check(err)
	return origTx
}

func progHash() hashing.HashValue {
	hash, err := hashing.HashValueFromBase58(scProgramHash)
	check(err)
	return hash
}

func activateSC(scAddress *address.Address) {
	for _, host := range config.CommitteeApi(committee()) {
		check(waspapi.ActivateSC(host, scAddress.String()))
	}
}

func postOriginTx(origTx *sctransaction.Transaction) {
	check(nodeapi.PostTransaction(config.GoshimmerApi(), origTx.Transaction))
}

func ownerAddress() address.Address {
	return wallet.Load().Address()
}

func committee() []int {
	return viper.GetIntSlice("fairroulette.committee")
}
