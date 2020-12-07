package sccmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/pflag"
)

func InitCommands(commands map[string]func([]string), flags *pflag.FlagSet) {
	commands["sc"] = cmd
}

var subcmds = map[string]func([]string){
	"deploy":     deployCmd,
	"activate":   activateCmd,
	"deactivate": deactivateCmd,
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

func parseIntList(s string) []int {
	committee := make([]int, 0)
	for _, ns := range strings.Split(s, ",") {
		n, err := strconv.Atoi(ns)
		log.Check(err)
		committee = append(committee, n)
	}
	return committee
}
