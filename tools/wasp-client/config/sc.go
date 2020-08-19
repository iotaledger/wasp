package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type SCConfig struct {
	ShortName string
	Flags     *pflag.FlagSet
	quorum    int
	committee []int
}

func (c *SCConfig) HookFlags() *pflag.FlagSet {
	c.Flags.IntVar(&c.quorum, c.ShortName+".quorum", 3, "quorum")
	c.Flags.IntSliceVar(&c.committee, c.ShortName+".committee", nil, "committee")
	return c.Flags
}

func (c *SCConfig) Committee() []int {
	if len(c.committee) > 0 {
		return c.committee
	}
	r := viper.GetIntSlice(c.ShortName + ".committee")
	if len(r) > 0 {
		return r
	}
	return []int{0, 1, 2, 3}
}

func (c *SCConfig) Quorum() int {
	return c.quorum
}

func (c *SCConfig) PrintUsage(s string) {
	fmt.Printf("Usage: %s %s %s\n", os.Args[0], c.ShortName, s)
}

func (c *SCConfig) HandleSetCmd(args []string) {
	if len(args) != 2 {
		c.PrintUsage("set <key> <value>")
		os.Exit(1)
	}
	Set(c.ShortName+"."+args[0], args[1])
}

func (c *SCConfig) usage(commands map[string]func([]string)) {
	cmdNames := make([]string, 0)
	for k := range commands {
		cmdNames = append(cmdNames, k)
	}

	c.PrintUsage(fmt.Sprintf("[options] [%s]", strings.Join(cmdNames, "|")))
	c.Flags.PrintDefaults()
	os.Exit(1)
}

func (c *SCConfig) HandleCmd(args []string, commands map[string]func([]string)) {
	if len(args) < 1 {
		c.usage(commands)
	}
	cmd, ok := commands[args[0]]
	if !ok {
		c.usage(commands)
	}
	cmd(args[1:])
}

func (c *SCConfig) SetAddress(address string) {
	SetSCAddress(c.ShortName, address)
}

func (c *SCConfig) Address() *address.Address {
	return GetSCAddress(c.ShortName)
}
