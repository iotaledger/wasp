package config

import (
	"fmt"

	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/packages/nodeconn"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/testutil/privtangle/privtangledefaults"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	ConfigPath        string
	WaitForCompletion bool
)

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		v := args[1]
		switch v {
		case "true":
			Set(args[0], true)
		case "false":
			Set(args[0], false)
		default:
			Set(args[0], v)
		}
	},
}

const (
	HostKindAPI     = "api"
	HostKindPeering = "peering"
	HostKindNanomsg = "nanomsg"
)

func Init(rootCmd *cobra.Command) {
	rootCmd.PersistentFlags().StringVarP(&ConfigPath, "config", "c", "wasp-cli.json", "path to wasp-cli.json")
	rootCmd.PersistentFlags().BoolVarP(&WaitForCompletion, "wait", "w", true, "wait for request completion")

	rootCmd.AddCommand(configSetCmd)
	rootCmd.AddCommand(checkVersionsCmd)

	// The first time parameters.L1() is called, it will be initialized with this function
	parameters.InitL1Lazy(func() {
		if viper.Get("l1.params") == nil {
			// get L1 params from node and save to config file
			log.Printf("Getting L1 params from node at %s...\n", L1APIAddress())
			L1Client() // this will call parameters.InitL1()
			Set("l1.params", parameters.L1())
		} else {
			// read L1 params from config file
			var params *parameters.L1Params
			err := viper.UnmarshalKey("l1.params", &params)
			log.Check(err)
			parameters.InitL1(params)
		}
	})
}

func Read() {
	viper.SetConfigFile(ConfigPath)
	_ = viper.ReadInConfig()
}

func L1APIAddress() string {
	host := viper.GetString("l1.apiAddress")
	if host != "" {
		return host
	}
	return fmt.Sprintf(
		"%s:%d",
		privtangledefaults.Host,
		privtangledefaults.BasePort+privtangledefaults.NodePortOffsetRestAPI,
	)
}

func L1FaucetAddress() string {
	address := viper.GetString("l1.faucetAddress")
	if address != "" {
		return address
	}
	return fmt.Sprintf(
		"%s:%d",
		privtangledefaults.Host,
		privtangledefaults.BasePort+privtangledefaults.NodePortOffsetFaucet,
	)
}

func L1Client() nodeconn.L1Client {
	log.Verbosef("using L1 API %s\n", L1APIAddress())

	return nodeconn.NewL1Client(
		nodeconn.L1Config{
			APIAddress:    L1APIAddress(),
			FaucetAddress: L1FaucetAddress(),
		},
		log.HiveLogger(),
	)
}

func GetToken() string {
	return viper.GetString("authentication.token")
}

func SetToken(token string) {
	Set("authentication.token", token)
}

func WaspClient(i ...int) *client.WaspClient {
	// TODO: add authentication for /adm
	log.Verbosef("using Wasp host %s\n", WaspAPI())
	L1Client() // this will fill parameters.L1() with data from the L1 node
	return client.NewWaspClient(WaspAPI(i...)).WithToken(GetToken())
}

func WaspAPI(i ...int) string {
	index := 0
	if len(i) > 0 {
		index = i[0]
	}
	r := viper.GetString("wasp." + HostKindAPI)
	if r != "" {
		return r
	}
	return committeeHost(HostKindAPI, index)
}

func WaspNanomsg(i ...int) string {
	index := 0
	if len(i) > 0 {
		index = i[0]
	}
	r := viper.GetString("wasp." + HostKindNanomsg)
	if r != "" {
		return r
	}
	return committeeHost(HostKindNanomsg, index)
}

func FindNodeBy(kind, v string) int {
	for i := 0; i < 100; i++ {
		if committeeHost(kind, i) == v {
			return i
		}
	}
	log.Fatalf("Cannot find node with %q = %q in configuration", kind, v)
	return 0
}

func CommitteeAPI(indices []int) []string {
	return committee(HostKindAPI, indices)
}

func CommitteePeering(indices []int) []string {
	return committee(HostKindPeering, indices)
}

func CommitteeNanomsg(indices []int) []string {
	return committee(HostKindNanomsg, indices)
}

func committee(kind string, indices []int) []string {
	hosts := make([]string, 0)
	for _, i := range indices {
		hosts = append(hosts, committeeHost(kind, i))
	}
	return hosts
}

func committeeConfigVar(kind string, i int) string {
	return fmt.Sprintf("wasp.%d.%s", i, kind)
}

func CommitteeAPIConfigVar(i int) string {
	return committeeConfigVar(HostKindAPI, i)
}

func CommitteePeeringConfigVar(i int) string {
	return committeeConfigVar(HostKindPeering, i)
}

func CommitteeNanomsgConfigVar(i int) string {
	return committeeConfigVar(HostKindNanomsg, i)
}

func committeeHost(kind string, i int) string {
	r := viper.GetString(committeeConfigVar(kind, i))
	if r != "" {
		return r
	}
	defaultPort := defaultWaspPort(kind, i)
	return fmt.Sprintf("127.0.0.1:%d", defaultPort)
}

func totalNumberOfWaspNodes() int {
	return len(viper.Sub("wasp").AllSettings())
}

func defaultWaspPort(kind string, i int) int {
	switch kind {
	case HostKindNanomsg:
		return 5550 + i
	case HostKindPeering:
		return 4000 + i
	case HostKindAPI:
		return 9090 + i
	}
	panic(fmt.Sprintf("no handler for kind %s", kind))
}

func Set(key string, value interface{}) {
	viper.Set(key, value)
	log.Check(viper.WriteConfig())
}
