package metrics

import (
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/spf13/cobra"
)

var metricsCmd = &cobra.Command{
	Use:   "metrics <component>",
	Short: "Show current value of collected metrics of some component",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		log.Check(cmd.Help())
	},
}

var chainAlias string

func Init(rootCmd *cobra.Command) {
	rootCmd.AddCommand(metricsCmd)

	metricsCmd.AddCommand(nodeconnMetricsCmd)
	metricsCmd.AddCommand(consensusMetricsCmd)
	metricsCmd.PersistentFlags().StringVarP(&chainAlias, "chain", "", "", "chain for which metrics should be displayed")
}
