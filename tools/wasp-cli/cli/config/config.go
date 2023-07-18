package config

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/spf13/viper"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp-wallet-sdk/types"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/testutil/privtangle/privtangledefaults"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/keychain"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

var (
	BaseDir           string
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

func locateBaseDir() string {
	homeDir, err := os.UserHomeDir()
	log.Check(err)

	_, err = os.Stat(homeDir)
	log.Check(err)

	baseDir := path.Join(homeDir, ".wasp-cli")
	_, err = os.Stat(baseDir)

	if err != nil {
		err = os.Mkdir(baseDir, os.ModePerm)
		log.Check(err)
	}

	BaseDir = baseDir
	return baseDir
}

func locateConfigFile() string {
	/*
		Searches for a wasp-cli.json at the current working directory,
		If not found, use the config file from the base dir (usually ~/.wasp-cli/wasp-cli.json)
	*/
	if ConfigPath == "" {
		cwd, err := os.Getwd()
		log.Check(err)

		_, err = os.Stat(path.Join(cwd, "wasp-cli.json"))
		if err == nil {
			ConfigPath = path.Join(cwd, "wasp-cli.json")
		} else {
			ConfigPath = path.Join(BaseDir, "wasp-cli.json")
		}
	}

	return ConfigPath
}

func Read() {
	locateBaseDir()
	locateConfigFile()

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

func GetToken(node string) string {
	token, err := keychain.GetJWTAuthToken(node)
	log.Check(err)
	return token
}

func SetToken(node, token string) {
	err := keychain.SetJWTAuthToken(node, token)
	log.Check(err)
}

func MustWaspAPIURL(nodeName string) string {
	apiAddress := WaspAPIURL(nodeName)
	if apiAddress == "" {
		log.Fatalf("wasp webapi not defined for node: %s", nodeName)
	}
	return apiAddress
}

func WaspAPIURL(nodeName string) string {
	return viper.GetString(fmt.Sprintf("wasp.%s", nodeName))
}

func NodeAPIURLs(nodeNames []string) []string {
	hosts := make([]string, 0)
	for _, nodeName := range nodeNames {
		hosts = append(hosts, MustWaspAPIURL(nodeName))
	}
	return hosts
}

func Set(key string, value interface{}) {
	viper.Set(key, value)
	log.Check(viper.WriteConfig())
}

func AddWaspNode(name, apiURL string) {
	Set("wasp."+name, apiURL)
}

func AddChain(name, chainID string) {
	Set("chains."+name, chainID)
}

func GetChain(name string) isc.ChainID {
	configChainID := viper.GetString("chains." + name)
	if configChainID == "" {
		log.Fatal(fmt.Sprintf("chain '%s' doesn't exist in config file", name))
	}
	networkPrefix, _, err := iotago.ParseBech32(configChainID)
	log.Check(err)

	if networkPrefix != parameters.L1().Protocol.Bech32HRP {
		err = fmt.Errorf("target network of the L1 node does not match the wasp-cli config")
	}
	log.Check(err)

	chainID, err := isc.ChainIDFromString(configChainID)
	log.Check(err)
	return chainID
}

func GetWalletSchemeString() string {
	return viper.GetString("wallet.scheme")
}

func SetWalletSchemeString(scheme string) {
	Set("wallet.scheme", scheme)
}

// GetSeedForMigration is used to migrate the seed of the config file to a certain wallet provider.
func GetSeedForMigration() string {
	return viper.GetString("wallet.seed")
}

func GetWalletLogLevel() types.ILoggerConfigLevelFilter {
	return types.ILoggerConfigLevelFilter(viper.GetString("wallet.loglevel"))
}

func SetWalletLogLevel(filter types.ILoggerConfigLevelFilter) {
	viper.Set("wallet.loglevel", string(filter))
}
