// +build evm

package chain

import (
	"encoding/base64"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/iotaledger/wasp/contracts/native/evmchain"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/spf13/cobra"
)

func init() {
	plugins = append(plugins, func(chainCmd *cobra.Command) {
		evmCmd := &cobra.Command{
			Use:   "evm <command>",
			Short: "Interact with EVM chains",
			Args:  cobra.NoArgs,
			Run: func(cmd *cobra.Command, args []string) {
				log.Check(cmd.Help())
			},
		}
		chainCmd.AddCommand(evmCmd)

		initEVMDeploy(evmCmd)
	})
}

func initEVMDeploy(evmCmd *cobra.Command) {
	var (
		name          string
		description   string
		alloc         []string
		genesisBase64 string
	)
	evmDeployCmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy the evmchain contract (i.e. create a new EVM chain)",
		Run: func(cmd *cobra.Command, args []string) {
			genesisAlloc := core.GenesisAlloc{}
			if len(alloc) != 0 && genesisBase64 != "" {
				log.Fatalf("--alloc and --alloc-bytes are mutually exclusive")
			}
			if len(alloc) == 0 && genesisBase64 == "" {
				log.Fatalf("One of --alloc and --alloc-bytes is mandatory")
			}
			if len(alloc) != 0 {
				for _, arg := range alloc {
					parts := strings.Split(arg, ":")
					addr := common.HexToAddress(parts[0])
					wei := big.NewInt(0)
					_, ok := wei.SetString(parts[1], 10)
					if !ok {
						log.Fatalf("cannot parse wei")
					}
					genesisAlloc[addr] = core.GenesisAccount{Balance: wei}
				}
			} else {
				b, err := base64.StdEncoding.DecodeString(genesisBase64)
				log.Check(err)
				genesisAlloc, err = evmchain.DecodeGenesisAlloc(b)
				log.Check(err)
			}

			deployContract(name, description, evmchain.Interface.ProgramHash, dict.Dict{
				evmchain.FieldGenesisAlloc: evmchain.EncodeGenesisAlloc(genesisAlloc),
			})
		},
	}
	evmCmd.AddCommand(evmDeployCmd)

	evmDeployCmd.Flags().StringVarP(&name, "name", "", "evmchain", "Contract name")
	evmDeployCmd.Flags().StringVarP(&description, "description", "", "EVM chain", "Contract description")
	evmDeployCmd.Flags().StringSliceVarP(&alloc, "alloc", "", nil, "Genesis allocation (format: <address>:<wei>,<address>:<wei>,...)")
	evmDeployCmd.Flags().StringVarP(&genesisBase64, "alloc-bytes", "", "", "Genesis allocation (base64-encoded)")
}
