package sccmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/iotaledger/wasp/tools/wwallet/config"
	"github.com/iotaledger/wasp/tools/wwallet/sc"
	"github.com/iotaledger/wasp/tools/wwallet/wallet"
)

func deployCmd(args []string) {
	if len(args) != 4 {
		deployUsage()
	}

	committee := parseIntList(args[0])
	quorum, err := strconv.Atoi(args[1])
	check(err)
	progHash := args[2]
	description := args[3]

	_, err = sc.Deploy(&sc.DeployParams{
		ProgramHash: progHash,
		Description: description,
		Quorum:      uint16(quorum),
		Committee:   committee,
		SigScheme:   wallet.Load().SignatureScheme(),
	})
	check(err)
}

func deployUsage() {
	fmt.Printf("Usage: %s sc deploy <committee> <quorum> <program-hash> <description>\n", os.Args[0])
	fmt.Printf("Example:\n")
	fmt.Printf("  %s set %s '%s'\n", os.Args[0], config.GoshimmerApiConfigVar(), config.GoshimmerApi())
	for i := 0; i < len(sc.DefaultCommittee); i++ {
		fmt.Printf("  %s set %s '%s'\n", os.Args[0], config.CommitteeApiConfigVar(i), config.CommitteeApi(sc.DefaultCommittee)[i])
		fmt.Printf("  %s set %s '%s'\n", os.Args[0], config.CommitteePeeringConfigVar(i), config.CommitteePeering(sc.DefaultCommittee)[i])
		fmt.Printf("  %s set %s '%s'\n", os.Args[0], config.CommitteeNanomsgConfigVar(i), config.CommitteeNanomsg(sc.DefaultCommittee)[i])
	}
	fmt.Printf("  %s --sc=fr sc deploy '0,1,2,3' 3 'FNT6snmmEM28duSg7cQomafbJ5fs596wtuNRn18wfaAz' 'FairRoulette'\n", os.Args[0])
	os.Exit(1)
}
