package wallet

import (
	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/tools/wasp-cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

func initRequestFundsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "request-funds",
		Short: "Request funds from the faucet",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			address := Load().Address()
			log.Check(config.L1Client().RequestFunds(address))

			model := &RequestFundsModel{
				Address: address.Bech32(parameters.L1().Protocol.Bech32HRP),
				Message: "success",
			}

			log.PrintCLIOutput(model)
		},
	}
}

type RequestFundsModel struct {
	Address string
	Message string
}

var _ log.CLIOutput = &RequestFundsModel{}

func (r *RequestFundsModel) AsText() (string, error) {
	template := `Request funds for address {{ .Address }} success`
	return log.ParseCLIOutputTemplate(r, template)
}
