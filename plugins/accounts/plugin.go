package accounts

import (
	"encoding/json"
	"github.com/iotaledger/hive.go/configuration"
	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/wasp/packages/jwt_auth"
	"github.com/iotaledger/wasp/packages/parameters"
)

type Account struct {
	Username string
	Password string
	Claims   []string
}

func (a *Account) GetTypedClaims() (*jwt_auth.AuthClaims, error) {
	claims := jwt_auth.AuthClaims{}
	fakeClaims := make(map[string]interface{})

	for _, v := range a.Claims {
		fakeClaims[v] = true
	}

	// TODO: Find a better solution for
	// Turning a list of strings into AuthClaims map by their json tag names
	enc, err := json.Marshal(fakeClaims)

	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(enc, &claims)

	if err != nil {
		return nil, err
	}

	return &claims, err
}

// PluginName is the name of the account plugin.
const PluginName = "Accounts"

var config *configuration.Configuration
var accounts []Account

func Init(_config *configuration.Configuration) *node.Plugin {
	config = _config
	return node.NewPlugin(PluginName, node.Enabled, configure, nil)
}

func configure(plugin *node.Plugin) {
	err := loadAccountsFromConfiguration()

	if err != nil {
		plugin.LogErrorf("Failed to pull accounts: {#err}")
	}
}

func loadAccountsFromConfiguration() error {
	err := config.Unmarshal(parameters.AccountsList, &accounts)

	if err != nil {
		accounts = make([]Account, 0)
	}

	return err
}

// TODO: Maybe add a DB connection later on, including functionality to remove/edit accounts?

func GetAccounts() *[]Account {
	return &accounts
}

func GetAccountByName(name string) *Account {
	for _, account := range accounts {
		if account.Username == name {
			return &account
		}
	}

	return nil
}
