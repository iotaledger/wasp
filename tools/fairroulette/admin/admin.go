package admin

import (
	"fmt"
	"os"
	"strconv"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	nodeapi "github.com/iotaledger/goshimmer/dapps/waspconn/packages/apilib"
	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/sctransaction/origin"
	"github.com/iotaledger/wasp/packages/vm/examples/fairroulette"
	"github.com/iotaledger/wasp/tools/fairroulette/config"
	"github.com/iotaledger/wasp/tools/fairroulette/util"
	"github.com/iotaledger/wasp/tools/fairroulette/wallet"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const scDescription = "FairRoulette smart contract"
const scProgramHash = fairroulette.ProgramHash

var quorumFlag int
var committeeFlag []int

func HookFlags() *pflag.FlagSet {
	flags := pflag.NewFlagSet("wallet init", pflag.ExitOnError)
	flags.IntVarP(&quorumFlag, "quorum", "t", 3, "quorum")
	flags.IntSliceVarP(&committeeFlag, "committee", "n", nil, "committee")
	return flags
}

func AdminCmd(args []string) {
	if len(args) < 1 {
		usage()
	}

	switch args[0] {
	case "init":
		initSC()

	case "set-period":
		if len(args) != 2 {
			fmt.Printf("Usage: %s admin set-period <seconds>\n", os.Args[0])
			os.Exit(1)
		}
		s, err := strconv.Atoi(args[1])
		check(err)
		setPeriod(s)

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
	fmt.Printf("Usage: %s admin [init|set-period <seconds>]\n", os.Args[0])
	os.Exit(1)
}

func initSC() {
	scAddress := genDKSets()
	origTx := createOriginTx(scAddress)
	color := balance.Color(origTx.ID())
	putScData(scAddress, &color)
	activateSC(scAddress)
	postOriginTx(origTx)
	fmt.Printf("Initialized %s\n", scDescription)
	fmt.Printf("SC Address: %s\n", scAddress.String())
	config.SetSCAddress(scAddress.String())
}

func genDKSets() *address.Address {
	scAddress, err := waspapi.GenerateNewDistributedKeySet(
		config.CommitteeApi(committee()),
		uint16(len(committee())),
		uint16(quorumFlag),
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
	ownerAddress := wallet.Load().Address()
	allOuts, err := nodeapi.GetAccountOutputs(config.GoshimmerApi(), &ownerAddress)
	check(err)
	origTx, err := origin.NewOriginTransaction(origin.NewOriginTransactionParams{
		Address:              *scAddress,
		OwnerSignatureScheme: wallet.Load().SignatureScheme(),
		AllInputs:            allOuts,
		InputColor:           balance.ColorIOTA,
		ProgramHash:          progHash(),
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
	if len(committeeFlag) > 0 {
		return committeeFlag
	}
	r := viper.GetIntSlice("fairroulette.committee")
	if len(r) > 0 {
		return r
	}
	return []int{0, 1, 2, 3}
}

func setPeriod(seconds int) {
	util.PostTransaction(&waspapi.RequestBlockJson{
		Address:     config.GetSCAddress().String(),
		RequestCode: fairroulette.RequestSetPlayPeriod,
		Vars: map[string]interface{}{
			fairroulette.ReqVarPlayPeriodSec: int64(seconds),
		},
	})
}
