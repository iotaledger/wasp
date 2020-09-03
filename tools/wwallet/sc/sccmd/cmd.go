package sccmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/iotaledger/wasp/tools/wwallet/config"
	"github.com/iotaledger/wasp/tools/wwallet/sc"
	"github.com/iotaledger/wasp/tools/wwallet/wallet"
	"github.com/spf13/pflag"
)

func InitCommands(commands map[string]func([]string), flags *pflag.FlagSet) {
	commands["sc"] = cmd
}

var subcmds = map[string]func([]string){
	"deploy": deployCmd,
}

func cmd(args []string) {
	if len(args) < 1 {
		usage()
	}
	subcmd, ok := subcmds[args[0]]
	if !ok {
		usage()
	}
	subcmd(args[1:])
}

func usage() {
	cmdNames := make([]string, 0)
	for k := range subcmds {
		cmdNames = append(cmdNames, k)
	}

	fmt.Printf("Usage: %s sc [%s]\n", os.Args[0], strings.Join(cmdNames, "|"))
	os.Exit(1)
}

func deployCmd(args []string) {
	if len(args) != 4 {
		deployUsage()
	}

	committee := parseCommittee(args[0])
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

func parseCommittee(s string) []int {
	committee := make([]int, 0)
	for _, ns := range strings.Split(s, ",") {
		n, err := strconv.Atoi(ns)
		check(err)
		committee = append(committee, n)
	}
	return committee
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
	fmt.Printf("  %s sc deploy '0,1,2,3' 3 'FNT6snmmEM28duSg7cQomafbJ5fs596wtuNRn18wfaAz' 'FairRoulette'\n", os.Args[0])
	os.Exit(1)
}

func check(err error) {
	if err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}
}
