package config

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

const defaultChain = "defaultchain"

var chainAlias string

func GetChainAlias() string {
	if chainAlias == "" {
		chainAlias = viper.GetString(defaultChain)
	}
	if chainAlias == "" {
		log.Fatal("No current chain. Call `chain deploy --chain=<alias>` or `set chain <alias>`")
	}
	return chainAlias
}

func SetCurrentChain(chainAlias string) {
	Set(defaultChain, chainAlias)
}

func InitAliasFlags(chainCmd *cobra.Command) {
	chainCmd.PersistentFlags().StringVarP(&chainAlias, defaultChain, "a", "", "chain alias")
}

func AddChainAlias(chainAlias, id string) {
	Set("chains."+chainAlias, id)
	SetCurrentChain(chainAlias)
}

func GetCurrentChainID() isc.ChainID {
	chainID, err := isc.ChainIDFromString(viper.GetString("chains." + GetChainAlias()))
	log.Check(err)
	return chainID
}
