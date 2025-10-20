package metrics

import (
	"github.com/spf13/cobra"
)

func initMetricsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "metrics <component>",
		Short: "Show current value of collected metrics of some component",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
}

var chainAlias string

func Init(rootCmd *cobra.Command) {
	metricsCmd := initMetricsCmd()
	rootCmd.AddCommand(metricsCmd)

	metricsCmd.AddCommand(initNodeconnMetricsCmd())
	metricsCmd.AddCommand(initConsensusMetricsCmd())
	metricsCmd.PersistentFlags().StringVarP(&chainAlias, "chain", "", "", "chain for which metrics should be displayed")
}
