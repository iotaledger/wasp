package config

import (
	"fmt"
	"os"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func init() {
	viper.SetDefault("goshimmer.api", "127.0.0.1:8080")

	viper.SetDefault("wasp.0.api", "127.0.0.1:9090")
	viper.SetDefault("wasp.1.api", "127.0.0.1:9091")
	viper.SetDefault("wasp.2.api", "127.0.0.1:9092")
	viper.SetDefault("wasp.3.api", "127.0.0.1:9093")

	viper.SetDefault("wasp.0.peering", "127.0.0.1:4000")
	viper.SetDefault("wasp.1.peering", "127.0.0.1:4001")
	viper.SetDefault("wasp.2.peering", "127.0.0.1:4002")
	viper.SetDefault("wasp.3.peering", "127.0.0.1:4003")
}

var configPath string
var Verbose bool

func HookFlags() *pflag.FlagSet {
	flags := pflag.NewFlagSet("wallet", pflag.ExitOnError)
	flags.StringVarP(&configPath, "config", "c", "fairroulette.json", "path to fairroulette.json")
	flags.BoolVarP(&Verbose, "verbose", "v", false, "verbose")
	return flags
}

func Read() {
	viper.SetConfigFile(configPath)
	viper.ReadInConfig()
}

func GoshimmerApi() string {
	return viper.GetString("goshimmer.api")
}

func WaspApi() string {
	return viper.GetString("wasp.0.api")
}

func CommitteeApi(indices []int) []string {
	return committee("api", indices)
}

func CommitteePeering(indices []int) []string {
	return committee("peering", indices)
}

func committee(kind string, indices []int) []string {
	hosts := make([]string, 0)
	for _, i := range indices {
		hosts = append(hosts, viper.GetString(fmt.Sprintf("wasp.%d.%s", i, kind)))
	}
	return hosts
}

func SetSCAddress(address string) {
	viper.SetDefault("fairroulette.address", address)
	viper.WriteConfig()
}

func GetSCAddress() address.Address {
	b58 := viper.GetString("fairroulette.address")
	if len(b58) == 0 {
		check(fmt.Errorf("call `set-address` first"))
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
