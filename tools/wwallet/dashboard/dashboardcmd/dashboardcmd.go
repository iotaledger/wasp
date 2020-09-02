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
)

func Cmd(args []string) {
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
	}
	if fa.Config.IsAvailable() {
		scs = append(scs, fadashboard.Dashboard())
	}
	if tr.Config.IsAvailable() {
		scs = append(scs, trdashboard.Dashboard())
	}
	if dwf.Config.IsAvailable() {
		scs = append(scs, dwfdashboard.Dashboard())
	}

	dashboard.StartServer(listenAddr, scs)
}
