package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"

	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/testutil/privtangle/privtangledefaults"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

var (
	ConfigPath        string
	WaitForCompletion bool
)

const (
	HostKindAPI     = "api"
	HostKindPeering = "peering"
	HostKindNanomsg = "nanomsg"

	l1ParamsKey          = "l1.params"
	l1ParamsTimestampKey = "l1.timestamp"
	l1ParamsExpiration   = 24 * time.Hour
)

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
	cliclients.L1Client() // this will call parameters.InitL1()
	Set(l1ParamsKey, parameters.L1NoLock())
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

func GetToken() string {
	return viper.GetString("authentication.token")
}

func SetToken(token string) {
	Set("authentication.token", token)
}

func MustWaspAPIURL(i ...int) string {
	apiAddress := WaspAPIURL(i...)
	if apiAddress == "" {
		panic("wasp webapi not defined")
	}
	return apiAddress
}

func WaspAPIURL(i ...int) string {
	index := 0
	if len(i) > 0 {
		index = i[0]
	}
	return viper.GetString(fmt.Sprintf("wasp.%d.%s", index, HostKindAPI))
}

func CommitteeAPIURL(indices []int) []string {
	return committeeHosts(HostKindAPI, indices)
}

func committeeHosts(kind string, indices []int) []string {
	hosts := make([]string, 0)
	for _, i := range indices {
		hosts = append(hosts, committeeHost(kind, i))
	}
	return hosts
}

func committeeConfigVar(kind string, i int) string {
	return fmt.Sprintf("wasp.%d.%s", i, kind)
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
	waspSubKey := viper.Sub("wasp")
	if waspSubKey == nil {
		return 0
	}

	return len(waspSubKey.AllSettings())
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
