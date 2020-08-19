package fairroulette

import (
	"fmt"
	"os"
	"strconv"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/sctransaction/origin"
	"github.com/iotaledger/wasp/packages/vm/examples/fairroulette"
	"github.com/iotaledger/wasp/tools/wasp-client/config"
	"github.com/iotaledger/wasp/tools/wasp-client/config/fr"
	"github.com/iotaledger/wasp/tools/wasp-client/wallet"
)

const scDescription = "FairRoulette smart contract"
const scProgramHash = fairroulette.ProgramHash

func adminCmd(args []string) {
	if len(args) < 1 {
		adminUsage()
	}

	switch args[0] {
	case "init":
		initSC()

	case "set-period":
		if len(args) != 2 {
			fr.Config.PrintUsage("admin set-period <seconds>")
			os.Exit(1)
		}
		s, err := strconv.Atoi(args[1])
		check(err)
		check(fr.Client().SetPeriod(s))

	default:
		adminUsage()
	}
}

func adminUsage() {
	fmt.Printf("Usage: %s fr admin [init|set-period <seconds>]\n", os.Args[0])
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
	fr.Config.SetAddress(scAddress.String())
}

func genDKSets() *address.Address {
	scAddress, err := waspapi.GenerateNewDistributedKeySetOld(
		config.CommitteeApi(fr.Config.Committee()),
		uint16(len(fr.Config.Committee())),
		uint16(fr.Config.Quorum()),
	)
	check(err)
	return scAddress
}

func putScData(scAddress *address.Address, color *balance.Color) {
	bootupData := registry.BootupData{
		Address:        *scAddress,
		Color:          *color,
		OwnerAddress:   ownerAddress(),
		CommitteeNodes: config.CommitteePeering(fr.Config.Committee()),
		AccessNodes:    []string{},
	}
	for _, host := range config.CommitteeApi(fr.Config.Committee()) {
		check(waspapi.PutSCData(host, bootupData))
	}
}

func createOriginTx(scAddress *address.Address) *sctransaction.Transaction {
	ownerAddress := wallet.Load().Address()
	allOuts, err := config.GoshimmerClient().GetAccountOutputs(&ownerAddress)
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
	for _, host := range config.CommitteeApi(fr.Config.Committee()) {
		check(waspapi.ActivateSC(host, scAddress.String()))
	}
}

func postOriginTx(origTx *sctransaction.Transaction) {
	check(config.GoshimmerClient().PostTransaction(origTx.Transaction))
}

func ownerAddress() address.Address {
	return wallet.Load().Address()
}
