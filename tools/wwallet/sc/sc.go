package sc

import (
	"fmt"
	"os"
	"strings"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/tools/wwallet/config"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Config struct {
	ShortName   string
	Description string
	ProgramHash string
	Flags       *pflag.FlagSet
	quorum      int
	committee   []int
}

func (c *Config) HookFlags() *pflag.FlagSet {
	c.Flags.IntVar(&c.quorum, c.ShortName+".quorum", 3, "quorum")
	c.Flags.IntSliceVar(&c.committee, c.ShortName+".committee", nil, "committee")
	return c.Flags
}

func (c *Config) Committee() []int {
	if len(c.committee) > 0 {
		return c.committee
	}
	r := viper.GetIntSlice(c.ShortName + ".committee")
	if len(r) > 0 {
		return r
	}
	return []int{0, 1, 2, 3}
}

func (c *Config) Quorum() uint16 {
	return uint16(c.quorum)
}

func (c *Config) PrintUsage(s string) {
	fmt.Printf("Usage: %s %s %s\n", os.Args[0], c.ShortName, s)
}

func (c *Config) HandleSetCmd(args []string) {
	if len(args) != 2 {
		c.PrintUsage("set <key> <value>")
		os.Exit(1)
	}
	config.Set(c.ShortName+"."+args[0], args[1])
}

func (c *Config) usage(commands map[string]func([]string)) {
	cmdNames := make([]string, 0)
	for k := range commands {
		cmdNames = append(cmdNames, k)
	}

	c.PrintUsage(fmt.Sprintf("[options] [%s]", strings.Join(cmdNames, "|")))
	c.Flags.PrintDefaults()
	os.Exit(1)
}

func (c *Config) HandleCmd(args []string, commands map[string]func([]string)) {
	if len(args) < 1 {
		c.usage(commands)
	}
	cmd, ok := commands[args[0]]
	if !ok {
		c.usage(commands)
	}
	cmd(args[1:])
}

func (c *Config) SetAddress(address string) {
	config.SetSCAddress(c.ShortName, address)
}

func (c *Config) Address() *address.Address {
	return config.GetSCAddress(c.ShortName)
}

func (c *Config) TryAddress() *address.Address {
	return config.TrySCAddress(c.ShortName)
}

func (c *Config) InitSC(sigScheme signaturescheme.SignatureScheme) error {
	scAddress, _, err := waspapi.CreateSC(waspapi.CreateSCParams{
		Node:                  config.GoshimmerClient(),
		CommitteeApiHosts:     config.CommitteeApi(c.Committee()),
		CommitteePeeringHosts: config.CommitteePeering(c.Committee()),
		AccessNodes:           []string{},
		N:                     uint16(len(c.Committee())),
		T:                     c.Quorum(),
		OwnerSigScheme:        sigScheme,
		ProgramHash:           c.progHash(),
	})
	if err != nil {
		return err
	}
	fmt.Printf("Initialized %s\n", c.Description)
	fmt.Printf("SC Addresses: %s\n", scAddress)
	c.SetAddress(scAddress.String())
	return nil
}

func (c *Config) progHash() hashing.HashValue {
	hash, err := hashing.HashValueFromBase58(c.ProgramHash)
	if err != nil {
		panic(err)
	}
	return hash
}
