package dashboardcmd

import (
	"fmt"
	"os"

	"github.com/iotaledger/wasp/tools/wwallet/dashboard"
	"github.com/iotaledger/wasp/tools/wwallet/sc/dwf"
	"github.com/iotaledger/wasp/tools/wwallet/sc/dwf/dwfdashboard"
	"github.com/iotaledger/wasp/tools/wwallet/sc/fa"
	"github.com/iotaledger/wasp/tools/wwallet/sc/fa/fadashboard"
	"github.com/iotaledger/wasp/tools/wwallet/sc/fr"
	"github.com/iotaledger/wasp/tools/wwallet/sc/fr/frdashboard"
	"github.com/iotaledger/wasp/tools/wwallet/sc/tr"
	"github.com/iotaledger/wasp/tools/wwallet/sc/tr/trdashboard"
	"github.com/spf13/pflag"
)

func InitCommands(commands map[string]func([]string), flags *pflag.FlagSet) {
	commands["dashboard"] = cmd
}

func cmd(args []string) {
	listenAddr := ":10000"
	if len(args) > 0 {
		if len(args) != 1 {
			fmt.Printf("Usage: %s dashboard [listen-address]\n", os.Args[0])
			os.Exit(1)
		}
		listenAddr = args[0]
	}

	scs := make([]dashboard.SCDashboard, 0)
	if fr.Config.IsAvailable() {
		scs = append(scs, frdashboard.Dashboard())
		fmt.Printf("FairRoulette: %s", fr.Config.Href())
	} else {
		fmt.Println("FairRoulette not available")
	}
	if fa.Config.IsAvailable() {
		scs = append(scs, fadashboard.Dashboard())
		fmt.Printf("FairAuction: %s", fa.Config.Href())
	} else {
		fmt.Println("FairAuction not available")
	}
	if tr.Config.IsAvailable() {
		scs = append(scs, trdashboard.Dashboard())
		fmt.Printf("TokenRegistry: %s", tr.Config.Href())
	} else {
		fmt.Println("TokenRegistry not available")
	}
	if dwf.Config.IsAvailable() {
		fmt.Printf("DonateWithFeedback: %s", dwf.Config.Href())
		scs = append(scs, dwfdashboard.Dashboard())
	} else {
		fmt.Println("DonateWithFeedback not available")
	}

	dashboard.StartServer(listenAddr, scs)
}
