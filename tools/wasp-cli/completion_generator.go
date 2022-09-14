package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

func completionCmd(name string) *cobra.Command {
	completionCmd := &cobra.Command{
		Use:   "gen_completion [bash|zsh|fish|powershell]",
		Short: "Generates shell autocompletion script. Run `wasp-cli gen_completion -h` for instructions.",
		Long: fmt.Sprintf(`wasp-cli needs to be installed (make install) for the to autocompletion work.

To load completions:

Bash:

  $ source <(%[1]s gen_completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ %[1]s gen_completion bash > /etc/bash_completion.d/%[1]s
  # macOS:
  $ %[1]s gen_completion bash > $(brew --prefix)/etc/bash_completion.d/%[1]s

Zsh:

  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:

  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ %[1]s gen_completion zsh > "${fpath[1]}/_%[1]s"

  # You will need to start a new shell for this setup to take effect.

fish:

  $ %[1]s gen_completion fish | source

  # To load completions for each session, execute once:
  $ %[1]s gen_completion fish > ~/.config/fish/completions/%[1]s.fish

PowerShell:

  PS> %[1]s gen_completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> %[1]s gen_completion powershell > %[1]s.ps1
  # and source this file from your PowerShell profile.
`, name),
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.ExactValidArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			switch args[0] {
			case "bash":
				log.Check(cmd.Root().GenBashCompletion(os.Stdout))
			case "zsh":
				log.Check(cmd.Root().GenZshCompletion(os.Stdout))
			case "fish":
				log.Check(cmd.Root().GenFishCompletion(os.Stdout, true))
			case "powershell":
				log.Check(cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout))
			}
		},
	}

	return completionCmd
}
