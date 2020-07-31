package config

import (
	"fmt"
	"os"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var configPath string
var Verbose bool

const (
	hostKindApi     = "api"
	hostKindPeering = "peering"
	hostKindNanomsg = "nanomsg"
)

func HookFlags() *pflag.FlagSet {
	flags := pflag.NewFlagSet("config", pflag.ExitOnError)
	flags.StringVarP(&configPath, "config", "c", "fairroulette.json", "path to fairroulette.json")
	flags.BoolVarP(&Verbose, "verbose", "v", false, "verbose")
	return flags
}

func Read() {
	viper.SetConfigFile(configPath)
	viper.ReadInConfig()
}

func GoshimmerApi() string {
	r := viper.GetString("goshimmer." + hostKindApi)
	if r != "" {
		return r
	}
	return "127.0.0.1:8080"
}

func WaspApi() string {
	r := viper.GetString("wasp." + hostKindApi)
	if r != "" {
		return r
	}
	return committeeHost(hostKindApi, 0)
}

func WaspNanomsg() string {
	r := viper.GetString("wasp." + hostKindApi)
	if r != "" {
		return r
	}
	return committeeHost(hostKindNanomsg, 0)
}

func CommitteeApi(indices []int) []string {
	return committee(hostKindApi, indices)
}

func CommitteePeering(indices []int) []string {
	return committee(hostKindPeering, indices)
}

func committee(kind string, indices []int) []string {
	hosts := make([]string, 0)
	for _, i := range indices {
		hosts = append(hosts, committeeHost(kind, i))
	}
	return hosts
}

func committeeHost(kind string, i int) string {
	r := viper.GetString(fmt.Sprintf("wasp.%d.%s", i, kind))
	if r != "" {
		return r
	}
	defaultPort := defaultWaspPort(kind, i)
	return fmt.Sprintf("127.0.0.1:%d", defaultPort)
}

func defaultWaspPort(kind string, i int) int {
	switch kind {
	case hostKindNanomsg:
		return 5550 + i
	case hostKindPeering:
		return 4000 + i
	case hostKindApi:
		return 9090 + i
	}
	panic(fmt.Sprintf("no handler for kind %s", kind))
}

func Set(key string, value interface{}) {
	viper.Set(key, value)
	viper.WriteConfig()
}

func SetSCAddress(address string) {
	Set("address", address)
}

func GetSCAddress() address.Address {
	b58 := viper.GetString("address")
	if len(b58) == 0 {
		check(fmt.Errorf("call `set-address` or `admin init` first"))
	}
	address, err := address.FromBase58(b58)
	check(err)
	return address
}

func check(err error) {
	if err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}
}
