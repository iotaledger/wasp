package verify

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/tools/wasp-cli/util"
	"github.com/spf13/cobra"
)

var ()

func initVerifyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "verify <command>",
		Short: "Verify code with Blockscout",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			log.Check(cmd.Help())
		},
	}
}

func parseImportMappings(mappings []string) (map[string]string, error) {
	result := make(map[string]string, len(mappings))
	for _, mapping := range mappings {
		if parts := strings.Split(mapping, "="); len(parts) != 2 {
			return nil, fmt.Errorf("each mapping must be a single key to a single real path (i.e. @iscmagic=packages/vm/core/evm/iscmagic)")
		} else {
			result[parts[0]] = parts[1]
		}
	}
	return result, nil
}

func initContractCmd() *cobra.Command {

	addressHash := new(common.Hash)
	var name string
	var compilerVersion string
	var optimization bool
	var constructorArguments string
	var autoDetectConstructorArguments bool
	var evmVersion string
	var optimizationRuns int
	var importRemaps []string

	cmd := &cobra.Command{
		Use:   "contract <blockscoutAPIAddress> <addressHash> <name>  <contractSourceCodeFilePath> [--args]",
		Short: "Verify a contract with blockscout",
		Args:  cobra.ExactArgs(4),
		Run: func(cmd *cobra.Command, args []string) {
			if blockscoutAPI, err := url.Parse(args[0]); err != nil {
				panic(err)
			} else if err := addressHash.UnmarshalText([]byte(args[1])); err != nil {
				panic(err)
			} else if name = args[2]; len(name) < 1 {
				panic(fmt.Errorf("the contract name must be more than 1 character long"))
			} else if _, err := os.Stat(args[3]); err != nil {
				panic(err)
			} else if remap, err := parseImportMappings(importRemaps); err != nil {
				panic(err)
			} else if err := VerifyContract(
				blockscoutAPI.String(),
				NewContract(
					addressHash,
					name,
					args[3],
					compilerVersion,
					constructorArguments,
					evmVersion,
					optimization,
					autoDetectConstructorArguments,
					optimizationRuns,
					remap,
				),
			); err != nil {
				panic(err)
			}
		},
	}

	cmd.Flags().StringVar(&compilerVersion, "compiler-version", util.SolcVersion, "version of the solidity compiler used to compile this contract")
	cmd.Flags().BoolVar(&optimization, "optimization", false, "Whether or not compiler optimizations were enabled")
	cmd.Flags().StringVar(&constructorArguments, "constructor-args", "", "The constructor argument data provided")
	cmd.Flags().BoolVar(&autoDetectConstructorArguments, "auto-detect-constructor-arguments", false, "Whether or not automatically detect constructor argument")
	cmd.Flags().StringVar(&evmVersion, "evm-version", "", "The EVM version for the contract")
	cmd.Flags().IntVar(&optimizationRuns, "optimization-runs", 0, "The number of optimization runs used during compilation")
	cmd.Flags().StringSliceVar(&importRemaps, "import-remapping", []string{}, "list of import mappings to re-assign (@iscmagic=packages/vm/core/evm/iscmagic,foo=/path/to/foo/source/files)")

	return cmd
}

func Init(rootCmd *cobra.Command) {
	verifyCmd := initVerifyCmd()
	rootCmd.AddCommand(verifyCmd)

	contractCmd := initContractCmd()

	verifyCmd.AddCommand(contractCmd)
}
