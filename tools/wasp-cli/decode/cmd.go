package decode

import (
	"encoding/json"
	"fmt"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/spf13/cobra"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain/statemanager/sm_gpa/sm_gpa_utils"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/subrealm"
	wasp_util "github.com/iotaledger/wasp/packages/util"
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
		Use:   "decode <type> <key> <type> ...",
		Short: "Decode the output of a contract function call",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			d := util.UnmarshalDict()

			if len(args) == 2 {
				ktype := args[0]
				vtype := args[1]

				for key, value := range d {
					skey := util.ValueToString(ktype, []byte(key))
					sval := util.ValueToString(vtype, value)
					log.Printf("%s: %s\n", skey, sval)
				}
				return
			}

			if len(args) < 3 || len(args)%3 != 0 {
				log.Check(cmd.Help())
				return
			}

			for i := 0; i < len(args)/2; i++ {
				ktype := args[i*2]
				skey := args[i*2+1]
				vtype := args[i*2+2]

				// chainID is only used to fallback user input, the decode command uses data directly from the server, it's okay to pass empty chainID
				key := kv.Key(util.ValueFromString(ktype, skey, isc.ChainID{}))
				val := d.Get(key)
				if val == nil {
					log.Printf("%s: <nil>\n", skey)
				} else {
					log.Printf("%s: %s\n", skey, util.ValueToString(vtype, val))
				}
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

			block, err := sm_gpa_utils.BlockFromFilePath(args[0])
			log.Check(err)

			fmt.Printf("Block Number: %v\n", block.StateIndex())
			fmt.Printf("L1 Commitment: %v\n", block.L1Commitment().String())

			blockInfos := make([]*blocklog.BlockInfo, 0)
			b := subrealm.NewReadOnly(block.MutationsReader(), kv.Key(blocklog.Contract.Hname().Bytes()))
			b.IterateKeys(blocklog.PrefixBlockRegistry, func(key kv.Key) bool {
				val2 := b.Get(key)
				info, blockErr := blocklog.BlockInfoFromBytes(val2)

				if blockErr == nil {
					blockInfos = append(blockInfos, info)
				}

				return true
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
			bytes, err := iotago.DecodeHex(args[0])
			log.Check(err)
			log.Printf(gas.MustFeePolicyFromBytes(bytes).String())
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
				ratio, err := wasp_util.Ratio32FromString(gasPerToken)
				log.Check(err)
				feePolicy.GasPerToken = ratio
			}

			if evmGasRatio != "" {
				ratio, err := wasp_util.Ratio32FromString(evmGasRatio)
				log.Check(err)
				feePolicy.EVMGasRatio = ratio
			}

			if validatorFeeShare <= 100 {
				feePolicy.ValidatorFeeShare = validatorFeeShare
			}

			log.Printf(iotago.EncodeHex(feePolicy.Bytes()))
		},
	}

	cmd.Flags().StringVar(&gasPerToken, "gasPerToken", "", "gas per token ratio (format: a:b)")
	cmd.Flags().StringVar(&evmGasRatio, "evmGasRatio", "", "evm gas ratio (format: a:b)")
	cmd.Flags().Uint8Var(&validatorFeeShare, "validatorFeeShare", 101, "validator fee share (between 0 and 100)")

	return cmd
}
