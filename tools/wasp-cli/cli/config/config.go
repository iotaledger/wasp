package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"

	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/testutil/privtangle/privtangledefaults"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

var (
	ConfigPath        string
	WaitForCompletion bool
)

const (
	l1ParamsKey          = "l1.params"
	l1ParamsTimestampKey = "l1.timestamp"
	l1ParamsExpiration   = 24 * time.Hour
)

func L1ParamsExpired() bool {
	if viper.Get(l1ParamsKey) == nil {
		return true
	}
	return viper.GetTime(l1ParamsTimestampKey).Add(l1ParamsExpiration).Before(time.Now())
}

func RefreshL1ParamsFromNode() {
	if log.VerboseFlag {
		log.Printf("Getting L1 params from node at %s...\n", L1APIAddress())
	}

	Set(l1ParamsKey, parameters.L1NoLock())
	Set(l1ParamsTimestampKey, time.Now())
}

func LoadL1ParamsFromConfig() {
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

func MustWaspAPIURL(nameOpt ...string) string {
	apiAddress := WaspAPIURL(nameOpt...)
	if apiAddress == "" {
		panic("wasp webapi not defined")
	}
	return apiAddress
}

func WaspAPIURL(nameOpt ...string) string {
	var nodeName string
	if len(nameOpt) > 0 {
		nodeName = nameOpt[0]
	} else {
		nodeName = MustGetDefaultWaspNode()
	}
	return viper.GetString(fmt.Sprintf("wasp.%s", nodeName))
}

func NodeAPIURLs(nodeNames []string) []string {
	hosts := make([]string, 0)
	for _, nodeName := range nodeNames {
		hosts = append(hosts, WaspAPIURL(nodeName))
	}
	return hosts
}

func Set(key string, value interface{}) {
	viper.Set(key, value)
	log.Check(viper.WriteConfig())
}
