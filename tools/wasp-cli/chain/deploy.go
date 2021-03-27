package chain

import (
	"os"

	"github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/tools/wasp-cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/tools/wasp-cli/wallet"
	"github.com/spf13/pflag"
)

var committee []int
var quorum int
var description string

func initDeployFlags(flags *pflag.FlagSet) {
	flags.IntSliceVarP(&committee, "committee", "", []int{0, 1, 2, 3}, "committee indices")
	flags.IntVarP(&quorum, "quorum", "", 3, "quorum")
	flags.StringVarP(&description, "description", "", "", "description")
}

func deployCmd(args []string) {
	alias := GetChainAlias()

	chainid, _, err := apilib.DeployChain(apilib.CreateChainParams{
		Node:                  config.GoshimmerClient(),
		CommitteeApiHosts:     config.CommitteeApi(committee),
		CommitteePeeringHosts: config.CommitteePeering(committee),
		N:                     uint16(len(committee)),
		T:                     uint16(quorum),
		OriginatorKeyPair:     wallet.Load().SignatureScheme(),
		Description:           description,
		Textout:               os.Stdout,
		Prefix:                "",
	})
	log.Check(err)

	AddChainAlias(alias, chainid.String())
}
