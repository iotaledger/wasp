package config

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/packages/l1connection"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/testutil/privtangle/privtangledefaults"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
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

var refreshL1ParamsCmd = &cobra.Command{
	Use:   "refresh-l1-params",
	Short: "Refresh L1 params from node",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		refreshL1ParamsFromNode()
	},
}

const (
	HostKindAPI     = "api"
	HostKindPeering = "peering"
	HostKindNanomsg = "nanomsg"

	l1ParamsKey          = "l1.params"
	l1ParamsTimestampKey = "l1.timestamp"
	l1ParamsExpiration   = 24 * time.Hour
)

func Init(rootCmd *cobra.Command) {
	rootCmd.PersistentFlags().StringVarP(&ConfigPath, "config", "c", "wasp-cli.json", "path to wasp-cli.json")
	rootCmd.PersistentFlags().BoolVarP(&WaitForCompletion, "wait", "w", true, "wait for request completion")

	rootCmd.AddCommand(configSetCmd)
	rootCmd.AddCommand(checkVersionsCmd)
	rootCmd.AddCommand(refreshL1ParamsCmd)

	// The first time parameters.L1() is called, it will be initialized with this function
	parameters.InitL1Lazy(func() {
		if l1ParamsExpired() {
			refreshL1ParamsFromNode()
		} else {
			loadL1ParamsFromConfig()
		}
	})
}

func l1ParamsExpired() bool {
	if viper.Get(l1ParamsKey) == nil {
		return true
	}
	return viper.GetTime(l1ParamsTimestampKey).Add(l1ParamsExpiration).Before(time.Now())
}

func refreshL1ParamsFromNode() {
	if log.VerboseFlag {
		log.Printf("Getting L1 params from node at %s...\n", L1APIAddress())
	}
	L1Client() // this will call parameters.InitL1()
	Set(l1ParamsKey, parameters.L1())
	Set(l1ParamsTimestampKey, time.Now())
}

func loadL1ParamsFromConfig() {
	// read L1 params from config file
	var params *parameters.L1Params
	err := viper.UnmarshalKey("l1.params", &params)
	log.Check(err)
	parameters.InitL1(params)
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

func L1INXAddress() string {
	host := viper.GetString("l1.inxAddress")
	if host != "" {
		return host
	}
	return fmt.Sprintf(
		"%s:%d",
		privtangledefaults.INXHost,
		privtangledefaults.BasePort+privtangledefaults.NodePortOffsetINX,
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

func L1Client() l1connection.Client {
	log.Verbosef("using L1 API %s\n", L1APIAddress())

	return l1connection.NewClient(
		l1connection.Config{
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
