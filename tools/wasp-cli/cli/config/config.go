// Package config handles configuration management for the wasp-cli tool,
// including reading, writing, and manipulating configuration settings.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"strings"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaconn"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/cli"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/cli/keychain"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/log"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/samber/lo"
)

var (
	BaseDir           string
	ConfigPath        string
	WaitForCompletion string
	PrettyPrintConfig bool

	Config = koanf.New(".")
)

const (
	DefaultWaitForCompletion = "0s"
)

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

	if err := Config.Load(file.Provider(ConfigPath), NewParser(PrettyPrintConfig)); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return
		}

		log.Check(err)
	}
}

func L1APIAddress() string {
	host := Config.String("l1.apiaddress")
	if host != "" {
		return host
	}
	return iotaconn.AlphanetEndpointURL
}

func L1FaucetAddress() string {
	address := Config.String("l1.faucetaddress")
	if address != "" {
		return address
	}
	return iotaconn.AlphanetFaucetURL
}

var keyChain keychain.KeyChain

func GetKeyChain() keychain.KeyChain {
	if keyChain == nil {
		if keychain.IsKeyChainAvailable() {
			keyChain = keychain.NewKeyChainZalando()
		} else {
			keyChain = keychain.NewKeyChainFile(BaseDir, cli.ReadPasswordFromStdin)
		}
	}

	return keyChain
}

func GetToken(node string) string {
	token, err := GetKeyChain().GetJWTAuthToken(node)
	log.Check(err)
	return token
}

func SetToken(node, token string) {
	err := GetKeyChain().SetJWTAuthToken(node, token)
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
	return Config.String(fmt.Sprintf("wasp.%s", strings.ToLower(nodeName)))
}

func NodeAPIURLs(nodeNames []string) []string {
	hosts := make([]string, 0)
	for _, nodeName := range nodeNames {
		hosts = append(hosts, MustWaspAPIURL(nodeName))
	}
	return hosts
}

func Set(key string, value interface{}) {
	log.Check(Config.Set(key, value))
	log.Check(WriteConfig())
}

func AddWaspNode(name, apiURL string) {
	Set("wasp."+name, apiURL)
}

func AddChain(name, chainID string) {
	Set("chains."+name, chainID)
}

func WriteConfig() error {
	b, err := Config.Marshal(NewParser(PrettyPrintConfig))
	if err != nil {
		return err
	}

	err = os.WriteFile(ConfigPath, b, 0o600)
	if err != nil {
		return err
	}
	return nil
}

func GetChain(name string) isc.ChainID {
	configChainID := Config.String("chains." + strings.ToLower(name))
	if configChainID == "" {
		log.Fatal(fmt.Sprintf("chain '%s' doesn't exist in config file", name))
	}
	_, err := cryptolib.NewAddressFromHexString(configChainID)
	log.Check(err)

	log.Check(err)

	chainID, err := isc.ChainIDFromString(configChainID)
	log.Check(err)
	return chainID
}

func GetPackageID() iotago.PackageID {
	configPackageID := Config.String("l1.packageid")
	if configPackageID == "" {
		log.Fatal(fmt.Sprintf("package id '%s' doesn't exist in config file", configPackageID))
	}

	packageIDParsed := lo.Must(iotago.PackageIDFromHex(configPackageID))
	return *packageIDParsed
}

func SetPackageID(id iotago.PackageID) {
	Set("l1.packageid", id.String())
}

func GetWalletProviderString() string {
	return Config.String("wallet.provider")
}

func SetWalletProviderString(provider string) {
	Set("wallet.provider", provider)
}

// GetSeedForMigration is used to migrate the seed of the config file to a certain wallet provider.
func GetSeedForMigration() string {
	return Config.String("wallet.seed")
}
func RemoveSeedForMigration() { log.Check(Config.Set("wallet.seed", "")) }

func GetAuthTokenForImport() map[string]string {
	stringMap := Config.Get("authentication.wasp")
	authTokenMap := map[string]string{}

	if mapData, ok := stringMap.(map[string]interface{}); ok {
		for k, v := range mapData {
			authTokenMap[k] = ""

			if authConfig, ok := v.(map[string]interface{}); ok {
				if token, ok := authConfig["token"].(string); ok {
					authTokenMap[k] = token
				}
			}
		}
	}

	return authTokenMap
}

func GetTestingSeed() string     { return Config.String("wallet.testing_seed") }
func SetTestingSeed(seed string) { log.Check(Config.Set("wallet.testing_seed", seed)) }

// Custom json parser for kaonf that supports formatting of JSON output
type JSON struct {
	prettyPrint bool
}

func NewParser(prettyPrint bool) *JSON {
	return &JSON{prettyPrint: prettyPrint}
}

func (p *JSON) Unmarshal(b []byte) (map[string]interface{}, error) {
	var out map[string]interface{}
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (p *JSON) Marshal(o map[string]interface{}) ([]byte, error) {
	if p.prettyPrint {
		return json.MarshalIndent(o, "", "  ")
	}
	return json.Marshal(o)
}
