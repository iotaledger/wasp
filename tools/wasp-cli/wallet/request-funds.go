package wallet

import (
	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/v2/tools/wasp-cli/util"

	"github.com/iotaledger/wasp/v2/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/cli/wallet"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/log"
)

func initRequestFundsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "request-funds",
		Short: "Request funds from the faucet",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			address := wallet.Load().Address()
			if err := cliclients.L1Client().RequestFundsFromFaucet(cmd.Context(), *address.AsIotaAddress()); err != nil {
				return err
			}

			model := &RequestFundsModel{
				Address: address.String(),
				Message: "success",
			}

			util.TryManageCoinsAmount(cmd.Context())

			log.PrintCLIOutput(model)
			return nil
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
