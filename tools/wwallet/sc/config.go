package sc

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/registry"
	"os"
	"strings"
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/tools/wwallet/config"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Config struct {
	ShortName        string
	Name             string
	ProgramHash      string
	Flags            *pflag.FlagSet
	committeeDefault []int
	quorumDefault    int
	bootupDataLoaded bool
	BootupData       registry.BootupData
}

func (c *Config) Href() string {
	return "/" + c.ShortName
}

func (c *Config) HookFlags() *pflag.FlagSet {
	c.Flags.IntVar(&c.quorumDefault, c.ShortName+".quorum", 0, "quorum (default 1,2,3,4)")
	c.Flags.IntSliceVar(&c.committeeDefault, c.ShortName+".committee", nil, "committee (default 3)")
	return c.Flags
}

var DefaultCommittee = []int{0, 1, 2, 3}

func (c *Config) SetCommittee(indexes []int) {
	config.Set(c.ShortName+".committee", indexes)
}

func (c *Config) Committee() []int {
	if len(c.committeeDefault) > 0 {
		return c.committeeDefault
	}
	r := viper.GetIntSlice(c.ShortName + ".committee")
	if len(r) > 0 {
		return r
	}
	return DefaultCommittee
}

func (c *Config) SetQuorum(n uint16) {
	config.Set(c.ShortName+".quorum", int(n))
}

func (c *Config) Quorum() uint16 {
	if c.quorumDefault != 0 {
		return uint16(c.quorumDefault)
	}
	q := viper.GetInt(c.ShortName + ".quorum")
	if q != 0 {
		return uint16(q)
	}
	return uint16(3)
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

func (c *Config) IsAvailable() bool {
	return config.TrySCAddress(c.ShortName) != nil
}

func (c *Config) Deploy(sigScheme signaturescheme.SignatureScheme) error {
	scAddress, err := Deploy(&DeployParams{
		Quorum:      c.Quorum(),
		Committee:   c.Committee(),
		Description: c.Name,
		ProgramHash: c.ProgramHash,
		SigScheme:   sigScheme,
	})
	if err == nil {
		c.SetAddress(scAddress.String())
		c.SetCommittee(c.Committee())
		c.SetQuorum(c.Quorum())
	}
	return err
}

type DeployParams struct {
	Quorum      uint16
	Committee   []int
	Description string
	ProgramHash string
	SigScheme   signaturescheme.SignatureScheme
}

func Deploy(params *DeployParams) (*address.Address, error) {
	scAddress, _, err := waspapi.CreateSC(waspapi.CreateSCParams{
		Node:                  config.GoshimmerClient(),
		CommitteeApiHosts:     config.CommitteeApi(params.Committee),
		CommitteePeeringHosts: config.CommitteePeering(params.Committee),
		AccessNodes:           []string{},
		N:                     uint16(len(params.Committee)),
		T:                     uint16(params.Quorum),
		OwnerSigScheme:        params.SigScheme,
		ProgramHash:           params.progHash(),
		Description:           params.Description,
		Textout:               os.Stdout,
		Prefix:                "[deploy] ",
	})
	if err != nil {
		return nil, err
	}
	err = waspapi.ActivateSCMulti(waspapi.ActivateSCParams{
		Addresses:         []*address.Address{scAddress},
		ApiHosts:          config.CommitteeApi(params.Committee),
		PublisherHosts:    config.CommitteeNanomsg(params.Committee),
		WaitForCompletion: config.WaitForConfirmation,
		Timeout:           30 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	fmt.Printf("Initialized %s smart contract\n", params.Description)
	fmt.Printf("SC Address: %s\n", scAddress)
	return scAddress, nil
}

func (p *DeployParams) progHash() hashing.HashValue {
	hash, err := hashing.HashValueFromBase58(p.ProgramHash)
	if err != nil {
		panic(err)
	}
	return hash
}

func LoadBootupData(cfg *Config) {
	if cfg.bootupDataLoaded {
		return
	}
	d, exists, err := waspapi.GetSCData(config.WaspApi(), cfg.Address())
	if err != nil || !exists {
		//fmt.Printf("++++++++++ GetSCData host = %s, addr = %s exists = %v err = %v\n",
		//	config.WaspApi(), cfg.Address(), exists, err)
		return
	}
	cfg.BootupData = *d
	cfg.bootupDataLoaded = true
	fmt.Printf("++++++++++ GetSCData %+v\n", cfg.BootupData)
}
