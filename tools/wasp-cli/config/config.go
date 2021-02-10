package config

import (
	"fmt"
	"os"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/client/level1"
	"github.com/iotaledger/wasp/client/level1/goshimmer"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var ConfigPath string
var WaitForCompletion bool

const (
	hostKindApi     = "api"
	hostKindPeering = "peering"
	hostKindNanomsg = "nanomsg"
)

func InitCommands(commands map[string]func([]string), flags *pflag.FlagSet) {
	commands["set"] = setCmd

	fs := pflag.NewFlagSet("config", pflag.ExitOnError)
	fs.StringVarP(&ConfigPath, "config", "c", "wasp-cli.json", "path to wasp-cli.json")
	fs.BoolVarP(&WaitForCompletion, "wait", "w", true, "wait for request completion")
	flags.AddFlagSet(fs)
}

func setCmd(args []string) {
	if len(args) != 2 {
		log.Usage("%s set <key> <value>\n", os.Args[0])
	}
	v := args[1]
	switch v {
	case "true":
		Set(args[0], true)
	case "false":
		Set(args[0], true)
	default:
		Set(args[0], v)
	}
}

func Read() {
	viper.SetConfigFile(ConfigPath)
	_ = viper.ReadInConfig()
}

func GoshimmerApiConfigVar() string {
	return "goshimmer." + hostKindApi
}

func GoshimmerApi() string {
	r := viper.GetString(GoshimmerApiConfigVar())
	if r != "" {
		return r
	}
	return "127.0.0.1:8080"
}

func Utxodb() bool {
	return viper.GetBool("utxodb")
}

func GoshimmerClient() level1.Level1Client {
	log.Verbose("using Goshimmer host %s\n", GoshimmerApi())
	if Utxodb() {
		log.Verbose("using utxodb\n")
		return testutil.NewGoshimmerUtxodbClient(GoshimmerApi())
	}
	return goshimmer.NewGoshimmerClient(GoshimmerApi())
}

func WaspClient() *client.WaspClient {
	log.Verbose("using Wasp host %s\n", WaspApi())
	return client.NewWaspClient(WaspApi())
}

func WaspApi() string {
	r := viper.GetString("wasp." + hostKindApi)
	if r != "" {
		return r
	}
	return committeeHost(hostKindApi, 0)
}

func WaspNanomsg() string {
	r := viper.GetString("wasp." + hostKindNanomsg)
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

func CommitteeNanomsg(indices []int) []string {
	return committee(hostKindNanomsg, indices)
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

func CommitteeApiConfigVar(i int) string {
	return committeeConfigVar(hostKindApi, i)
}

func CommitteePeeringConfigVar(i int) string {
	return committeeConfigVar(hostKindPeering, i)
}

func CommitteeNanomsgConfigVar(i int) string {
	return committeeConfigVar(hostKindNanomsg, i)
}

func committeeHost(kind string, i int) string {
	r := viper.GetString(committeeConfigVar(kind, i))
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
	log.Check(viper.WriteConfig())
}

func TrySCAddress(scAlias string) *address.Address {
	b58 := viper.GetString("sc." + scAlias + ".address")
	if len(b58) == 0 {
		return nil
	}
	address, err := address.FromBase58(b58)
	log.Check(err)
	return &address
}

func GetSCAddress(scAlias string) *address.Address {
	address := TrySCAddress(scAlias)
	if address == nil {
		log.Fatal("call `%s set sc.%s.address` or deploy a contract first", os.Args[0], scAlias)
	}
	return address
}
