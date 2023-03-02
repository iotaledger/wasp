package decode

import (
	"encoding/json"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/spf13/cobra"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	wasp_util "github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/gas"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/tools/wasp-cli/util"
)

func Init(rootCmd *cobra.Command) {
	rootCmd.AddCommand(initDecodeCmd())
	rootCmd.AddCommand(initDecodeMetadataCmd())
	rootCmd.AddCommand(initDecodeGasFeePolicy())
	rootCmd.AddCommand(initEncodeGasFeePolicy())
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

				key := kv.Key(util.ValueFromString(ktype, skey))
				val := d.MustGet(key)
				if val == nil {
					log.Printf("%s: <nil>\n", skey)
				} else {
					log.Printf("%s: %s\n", skey, util.ValueToString(vtype, val))
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
		Use:   "decode-gaspolicy <0x...>",
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
		Use:   "encode-gaspolicy",
		Short: "Translates metadata from Hex to a humanly-readable format",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			gasPolicy := gas.DefaultFeePolicy()

			if gasPerToken != "" {
				ratio, err := wasp_util.Ratio32FromString(gasPerToken)
				log.Check(err)
				gasPolicy.GasPerToken = ratio
			}

			if evmGasRatio != "" {
				ratio, err := wasp_util.Ratio32FromString(evmGasRatio)
				log.Check(err)
				gasPolicy.EVMGasRatio = ratio
			}

			if validatorFeeShare <= 100 {
				gasPolicy.ValidatorFeeShare = validatorFeeShare
			}

			log.Printf(iotago.EncodeHex(gasPolicy.Bytes()))
		},
	}

	cmd.Flags().StringVar(&gasPerToken, "gasPerToken", "", "gas per token ratio (format: a:b)")
	cmd.Flags().StringVar(&evmGasRatio, "evmGasRatio", "", "evm gas ratio (format: a:b)")
	cmd.Flags().Uint8Var(&validatorFeeShare, "validatorFeeShare", 101, "validator fee share (between 0 and 100)")

	return cmd
}
