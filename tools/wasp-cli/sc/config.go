// +build ignore

package sc

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/iotaledger/wasp/packages/registry_pkg/chainrecord"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/client/chainclient"
	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/tools/wasp-cli/chain"
	"github.com/iotaledger/wasp/tools/wasp-cli/config"
	"github.com/spf13/viper"
)

type Config struct {
	ShortName   string
	Name        string
	ProgramHash string

	chainRecord *chainrecord.ChainRecord
}

func (c *Config) MakeClient(sigScheme *ed25519.KeyPair) *chainclient.Client {
	client := chainclient.New(
		config.GoshimmerClient(),
		client.NewWaspClient(config.WaspApi()),
		chain.GetCurrentChainID(),
		sigScheme,
	)
	return client
}

func (c *Config) Alias() string {
	if config.SCAlias != "" {
		return config.SCAlias
	}
	if c.ShortName != "" {
		return c.ShortName
	}
	panic("Which smart contract? (--sc=<alias> is required)")
}

func (c *Config) Href() string {
	return "/" + c.ShortName
}

var DefaultCommittee = []int{0, 1, 2, 3}

func (c *Config) SetCommittee(indexes []int) {
	config.Set("sc."+c.Alias()+".committee", indexes)
}

func (c *Config) Committee() []int {
	r := viper.GetIntSlice("sc." + c.Alias() + ".committee")
	if len(r) > 0 {
		return r
	}
	return DefaultCommittee
}

func (c *Config) SetQuorum(n uint16) {
	config.Set("sc."+c.Alias()+".quorum", int(n))
}

func (c *Config) Quorum() uint16 {
	q := viper.GetInt("sc." + c.Alias() + ".quorum")
	if q != 0 {
		return uint16(q)
	}
	return uint16(3)
}

func (c *Config) usage(s string) {
	fmt.Usage("%s %s %s\n", os.Args[0], c.ShortName, s)
}

func (c *Config) HandleSetCmd(args []string) {
	if len(args) != 2 {
		c.usage("set <key> <value>")
	}
	config.Set("sc."+c.Alias()+"."+args[0], args[1])
}

func (c *Config) usage(commands map[string]func([]string)) {
	cmdNames := make([]string, 0)
	for k := range commands {
		cmdNames = append(cmdNames, k)
	}

	c.usage(fmt.Sprintf("[options] [%s]", strings.Join(cmdNames, "|")))
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
	config.SetSCAddress(c.Alias(), address)
}

func (c *Config) Address() ledgerstate.Address {
	return config.GetSCAddress(c.Alias())
}

func (c *Config) IsAvailable() bool {
	return config.TrySCAddress(c.Alias()) != nil
}

func (c *Config) Deploy(sigScheme *ed25519.KeyPair) error {
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
	SigScheme   *ed25519.KeyPair
}

func Deploy(params *DeployParams) (ledgerstate.Address, error) {
	scAddress, _, err := waspapi.DeployChainWithDKG(waspapi.CreateChainParams{
		Node:                  config.GoshimmerClient(),
		CommitteeApiHosts:     config.CommitteeApi(params.Committee),
		CommitteePeeringHosts: config.CommitteePeering(params.Committee),
		N:                     uint16(len(params.Committee)),
		T:                     uint16(params.Quorum),
		OriginatorKeyPair:     params.SigScheme,
		ProgramHash:           params.progHash(),
		Description:           params.Description,
		Textout:               os.Stdout,
		Prefix:                "[deploy] ",
	})
	if err != nil {
		return nil, err
	}
	err = waspapi.ActivateChain(waspapi.ActivateChainParams{
		ChainID:           []ledgerstate.Address{scAddress},
		ApiHosts:          config.CommitteeApi(params.Committee),
		PublisherHosts:    config.CommitteeNanomsg(params.Committee),
		WaitForCompletion: config.WaitForCompletion,
		Timeout:           30 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	fmt.Printf("Initialized %s smart contract\n", params.Description)
	fmt.Printf("SC Target: %s\n", scAddress)

	if config.SCAlias != "" {
		c := Config{
			ProgramHash: params.ProgramHash,
		}
		c.SetAddress(scAddress.String())
		c.SetCommittee(params.Committee)
		c.SetQuorum(params.Quorum)
	}

	return scAddress, nil
}

func (p *DeployParams) progHash() hashing.HashValue {
	hash, err := hashing.HashValueFromBase58(p.ProgramHash)
	if err != nil {
		panic(err)
	}
	return hash
}

func (c *Config) ChainRecord() *chainrecord.ChainRecord {
	if c.chainRecord != nil {
		return c.chainRecord
	}
	d, err := c.MakeClient(nil).GetChainRecord()
	if err != nil {
		panic(fmt.Sprintf("GetChainRecord failed: host = %s, addr = %s err = %v\n",
			config.WaspApi(), c.Address(), err))
	}
	c.chainRecord = d
	return c.chainRecord
}
