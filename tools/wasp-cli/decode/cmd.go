package decode

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/packages/chain/statemanager/gpa/utils"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	wasputil "github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/gas"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/tools/wasp-cli/util"
)

func Init(rootCmd *cobra.Command) {
	rootCmd.AddCommand(initDecodeCmd())
	rootCmd.AddCommand(initDecodeMetadataCmd())
	rootCmd.AddCommand(initDecodeGasFeePolicy())
	rootCmd.AddCommand(initEncodeGasFeePolicy())
	rootCmd.AddCommand(initDecodeWALCmd())
}

func initDecodeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "decode <type> <type> ...",
		Short: "Decode the output of a contract function call",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, cmdArgs []string) {
			callResults := util.ReadCallResultsAsJSON()

			if len(cmdArgs) < 1 {
				log.Check(cmd.Help())
				return
			}

			if len(callResults) != len(cmdArgs) {
				log.Printf("Number of provided result types does not match number of results: types = %v, results = %v\n",
					len(cmdArgs), len(callResults))
				os.Exit(1)
				return
			}

			for i := 0; i < len(cmdArgs); i++ {
				vtype := cmdArgs[i]
				val := util.ValueToString(vtype, callResults[i])
				log.Printf("[%v]: %s\n", i, val)
			}
		},
	}
}

func initDecodeWALCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "decode-wal <path>",
		Short: "Parses and dumps a WAL file",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Reading WAL file '%s'\n", args[0])

			block, err := utils.BlockFromFilePath(args[0])
			log.Check(err)

			fmt.Printf("Block Number: %v\n", block.StateIndex())
			fmt.Printf("L1 Commitment: %v\n", block.L1Commitment().String())

			blockInfos := make([]*blocklog.BlockInfo, 0)
			blocklog.NewStateReaderFromBlockMutations(block).IterateBlockRegistryPrefix(func(bi *blocklog.BlockInfo) {
				blockInfos = append(blockInfos, bi)
			})

			fmt.Printf("Found BlockInfos:\n\n")
			if len(blockInfos) == 0 {
				fmt.Println("None")
			} else {
				for i, info := range blockInfos {
					fmt.Printf("%v:\n", i)
					fmt.Println(info)
					fmt.Println("")
				}
			}

			receipts, err := blocklog.RequestReceiptsFromBlock(block)
			fmt.Printf("\nRequests:\n\n")

			if err != nil {
				fmt.Println("Failed to decode receipts")
				fmt.Println(err)
			} else {
				if len(receipts) == 0 {
					fmt.Printf("No requests\n")
				} else {
					for i, receipt := range receipts {
						fmt.Printf("%v:\n", i)
						fmt.Printf("%v\n", receipt.String())
					}
					fmt.Printf("\n\n")
				}
			}
		},
	}
}

func initDecodeMetadataCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "decode-metadata <0x...>",
		Short: "Translates metadata from Hex to a humanly-readable format",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			metadata, err := isc.RequestMetadataFromBytes(hexutil.MustDecode(args[0]))
			log.Check(err)
			jsonBytes, err := json.MarshalIndent(metadata, "", "  ")
			log.Check(err)
			log.Printf("%s\n", jsonBytes)
		},
	}
}

func initDecodeGasFeePolicy() *cobra.Command {
	return &cobra.Command{
		Use:   "decode-feepolicy <0x...>",
		Short: "Translates gas fee policy from Hex to a humanly-readable format",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			bytes, err := cryptolib.DecodeHex(args[0])
			log.Check(err)
			log.Printf("%v", gas.MustFeePolicyFromBytes(bytes).String())
		},
	}
}

func initEncodeGasFeePolicy() *cobra.Command {
	var (
		gasPerToken       string
		evmGasRatio       string
		validatorFeeShare uint8
	)

	cmd := &cobra.Command{
		Use:   "encode-feepolicy",
		Short: "Translates metadata from Hex to a humanly-readable format",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			feePolicy := gas.DefaultFeePolicy()

			if gasPerToken != "" {
				ratio, err := wasputil.Ratio32FromString(gasPerToken)
				log.Check(err)
				feePolicy.GasPerToken = ratio
			}

			if evmGasRatio != "" {
				ratio, err := wasputil.Ratio32FromString(evmGasRatio)
				log.Check(err)
				feePolicy.EVMGasRatio = ratio
			}

			if validatorFeeShare <= 100 {
				feePolicy.ValidatorFeeShare = validatorFeeShare
			}

			log.Printf("%s", cryptolib.EncodeHex(feePolicy.Bytes()))
		},
	}

	cmd.Flags().StringVar(&gasPerToken, "gasPerToken", "", "gas per token ratio (format: a:b)")
	cmd.Flags().StringVar(&evmGasRatio, "evmGasRatio", "", "evm gas ratio (format: a:b)")
	cmd.Flags().Uint8Var(&validatorFeeShare, "validatorFeeShare", 101, "validator fee share (between 0 and 100)")

	return cmd
}
