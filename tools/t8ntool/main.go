package main

import (
	"log"
	"os"

	t8ntool "github.com/iotaledger/wasp/tools/t8ntool/pkg"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "evm",
		Usage: "a t8n interface which pretends to be evm",
		// FIXME a fake version to match the evm versioning scheme
		// use "github.com/ethereum/go-ethereum/version" to set the version from the go-ethereum version package
		Version: "1.15.11-stable",
		Commands: []*cli.Command{
			{
				Name:    "transition",
				Aliases: []string{"t8n"},
				Usage:   "Executes a full state transition",
				Action:  t8ntool.Transition,
				Flags: []cli.Flag{
					t8ntool.TraceFlag,
					t8ntool.TraceTracerFlag,
					t8ntool.TraceTracerConfigFlag,
					t8ntool.TraceEnableMemoryFlag,
					t8ntool.TraceDisableStackFlag,
					t8ntool.TraceEnableReturnDataFlag,
					t8ntool.TraceEnableCallFramesFlag,
					t8ntool.OutputBasedir,
					t8ntool.OutputAllocFlag,
					t8ntool.OutputResultFlag,
					t8ntool.OutputBodyFlag,
					t8ntool.InputAllocFlag,
					t8ntool.InputEnvFlag,
					t8ntool.InputTxsFlag,
					t8ntool.ForknameFlag,
					t8ntool.ChainIDFlag,
					t8ntool.RewardFlag,
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
