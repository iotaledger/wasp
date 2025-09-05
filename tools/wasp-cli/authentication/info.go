package authentication

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/v2/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/waspcmd"
)

type AuthInfoOutput struct {
	AuthenticationMethod string
	AuthenticationURL    string
}

var _ log.CLIOutput = &AuthInfoOutput{}

func (l *AuthInfoOutput) AsText() (string, error) {
	template := `Authentication Method: {{ .AuthenticationMethod }}
Authentication URL: {{ .AuthenticationURL }}`
	return log.ParseCLIOutputTemplate(l, template)
}

func initInfoCmd() *cobra.Command {
	var node string
	cmd := &cobra.Command{
		Use:   "info",
		Short: "Receive information about the authentication methods",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Auth is currently not inside Swagger, so this is a temporary change
			ctx := context.Background()
			var err error
			node, err = waspcmd.DefaultWaspNodeFallback(node)
			if err != nil {
				return err
			}
			client := cliclients.WaspClientWithVersionCheck(ctx, node)
			authInfo, _, err := client.AuthAPI.AuthInfo(ctx).Execute()

			if err != nil {
				return err
			}

			log.PrintCLIOutput(&AuthInfoOutput{
				AuthenticationMethod: authInfo.Scheme,
				AuthenticationURL:    authInfo.AuthURL,
			})
			return nil
		},
	}
	waspcmd.WithWaspNodeFlag(cmd, &node)
	return cmd
}
