package chain

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/tools/wasp-cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var chainAlias string

func GetChainAlias() string {
	if chainAlias == "" {
		chainAlias = viper.GetString("chain")
	}
	if chainAlias == "" {
		log.Fatalf("No current chain. Call `chain deploy --chain=<alias>` or `set chain <alias>`")
	}
	return chainAlias
}

func SetCurrentChain(chainAlias string) {
	config.Set("chain", chainAlias)
}

func initAliasFlags(chainCmd *cobra.Command) {
	chainCmd.PersistentFlags().StringVarP(&chainAlias, "chain", "a", "", "chain alias")
}

func AddChainAlias(chainAlias, id string) {
	config.Set("chains."+chainAlias, id)
	SetCurrentChain(chainAlias)
}

func GetCurrentChainID() *isc.ChainID {
	chid, err := isc.ChainIDFromString(viper.GetString("chains." + GetChainAlias()))
	log.Check(err)
	return chid
}
